require 'sinatra/base'
require 'mysql2-cs-bind'
require 'rack-flash'
require 'shellwords'
require 'rack/session/dalli'
require 'fileutils'
require 'openssl'

module Isuconp
  class App < Sinatra::Base
    use Rack::Session::Dalli, autofix_keys: true, secret: ENV['ISUCONP_SESSION_SECRET'] || 'sendagaya', memcache_server: ENV['ISUCONP_MEMCACHED_ADDRESS'] || 'localhost:11211'
    use Rack::Flash
    set :public_folder, File.expand_path('../../public', __FILE__)

    UPLOAD_LIMIT = 10 * 1024 * 1024 # 10mb

    POSTS_PER_PAGE = 20

    IMAGE_DIR = File.expand_path('../../public/image', __FILE__)

    helpers do
      def config
        @config ||= {
          db: {
            host: ENV['ISUCONP_DB_HOST'] || 'localhost',
            port: ENV['ISUCONP_DB_PORT'] && ENV['ISUCONP_DB_PORT'].to_i,
            username: ENV['ISUCONP_DB_USER'] || 'root',
            password: ENV['ISUCONP_DB_PASSWORD'],
            database: ENV['ISUCONP_DB_NAME'] || 'isuconp',
          },
        }
      end

      def memcached
        return Thread.current[:memcached] if Thread.current[:memcached]
        client = Dalli::Client.new(ENV['ISUCONP_MEMCACHED_ADDRESS'] || 'localhost:11211', {})
        Thread.current[:memcached] = client
        client
      end

      def db
        return Thread.current[:isuconp_db] if Thread.current[:isuconp_db]
        client = Mysql2::Client.new(
          host: config[:db][:host],
          port: config[:db][:port],
          username: config[:db][:username],
          password: config[:db][:password],
          database: config[:db][:database],
          encoding: 'utf8mb4',
          reconnect: true,
        )
        client.query_options.merge!(symbolize_keys: true, database_timezone: :local, application_timezone: :local)
        Thread.current[:isuconp_db] = client
        client
      end

      def db_initialize
        sql = []
        sql << 'DELETE FROM users WHERE id > 1000'
        sql << 'DELETE FROM posts WHERE id > 10000'
        sql << 'DELETE FROM comments WHERE id > 100000'
        sql << 'UPDATE users SET del_flg = 0'
        sql << 'UPDATE users SET del_flg = 1 WHERE id % 50 = 0'
        sql.each do |s|
          db.prepare(s).execute
        end
      end

      def try_login(account_name, password)
        user = db.xquery('SELECT * FROM users WHERE account_name = ? AND del_flg = 0', account_name).first

        if user && calculate_passhash(user[:account_name], password) == user[:passhash]
          return user
        elsif user
          return nil
        else
          return nil
        end
      end

      def validate_user(account_name, password)
        if !(/\A[0-9a-zA-Z_]{3,}\z/.match(account_name) && /\A[0-9a-zA-Z_]{6,}\z/.match(password))
          return false
        end

        return true
      end

      def digest(src)
        return OpenSSL::Digest::SHA512.hexdigest(src)
      end

      def calculate_salt(account_name)
        digest account_name
      end

      def calculate_passhash(account_name, password)
        digest "#{password}:#{calculate_salt(account_name)}"
      end

      def get_session_user()
        if session[:user]
          db.xquery('SELECT * FROM `users` WHERE `id` = ?',
            session[:user][:id]
          ).first
        else
          nil
        end
      end

      def make_posts(results, all_comments: false)
        posts = []
        # posts.idをあらかじめ取り出してキャッシュのキーを一覧にする
        count_keys = results.to_a.map{|post| "comments.#{post[:id]}.count"}
        comments_keys = results.to_a.map{|post| "comments.#{post[:id]}.#{all_comments.to_s}"}

        # get_multiで複数のキーを一度に取得する
        cached_counts = memcached.get_multi(count_keys)
        cached_comments = memcached.get_multi(comments_keys)

        results.to_a.each do |post|
          if cached_counts["comments.#{post[:id]}.count"]
            # 取得済みのキャッシュがあればそれを使う
            post[:comment_count] = cached_counts["comments.#{post[:id]}.count"].to_i
          else
            # 存在しなかったらMySQLにクエリ
            post[:comment_count] = db.xquery('SELECT COUNT(*) AS `count` FROM `comments` WHERE `post_id` = ?',
              post[:id]
            ).first[:count]
            # memcachedにset(TTL 10s)
            memcached.set("comments.#{post[:id]}.count", post[:comment_count], 10)
          end

          if cached_comments["comments.#{post[:id]}.#{all_comments.to_s}"]
            # 取得済みのキャッシュがあればそれを使う
            post[:comments] = cached_comments["comments.#{post[:id]}.#{all_comments.to_s}"]
          else
            # 存在しなかったらMySQLにクエリ JOINで1クエリで取得する
            query = 'SELECT c.`comment`, c.`created_at`, u.`account_name`
                    FROM `comments` c JOIN `users` u
                    ON c.`user_id`=u.`id`
                    WHERE c.`post_id` = ? ORDER BY c.`created_at` DESC'
            unless all_comments
              query += ' LIMIT 3'
            end
            comments = db.xquery(query, post[:id]).to_a
            comments.each do |comment|
              comment[:user] = {
                account_name: comment[:account_name]
              }
            end
            post[:comments] = comments.reverse
            # memcachedにset(TTL 10s)
            memcached.set("comments.#{post[:id]}.#{all_comments.to_s}", post[:comments], 10)
          end

          post[:user] = {
            account_name: post[:account_name],
          }

          posts.push(post)
        end

        posts
      end

      def image_url(post)
        ext = ""
        if post[:mime] == "image/jpeg"
          ext = ".jpg"
        elsif post[:mime] == "image/png"
          ext = ".png"
        elsif post[:mime] == "image/gif"
          ext = ".gif"
        end

        "/image/#{post[:id]}#{ext}"
      end
    end

    get '/initialize' do
      db_initialize
      return 200
    end

    get '/login' do
      if get_session_user()
        redirect '/', 302
      end
      erb :login, layout: :layout, locals: { me: nil }
    end

    post '/login' do
      if get_session_user()
        redirect '/', 302
      end

      user = try_login(params['account_name'], params['password'])
      if user
        session[:user] = {
          id: user[:id]
        }
        session[:csrf_token] = SecureRandom.hex(16)
        redirect '/', 302
      else
        flash[:notice] = 'アカウント名かパスワードが間違っています'
        redirect '/login', 302
      end
    end

    get '/register' do
      if get_session_user()
        redirect '/', 302
      end
      erb :register, layout: :layout, locals: { me: nil }
    end

    post '/register' do
      if get_session_user()
        redirect '/', 302
      end

      account_name = params['account_name']
      password = params['password']

      validated = validate_user(account_name, password)
      if !validated
        flash[:notice] = 'アカウント名は3文字以上、パスワードは6文字以上である必要があります'
        redirect '/register', 302
        return
      end

      user = db.xquery('SELECT 1 FROM users WHERE `account_name` = ?', account_name).first
      if user
        flash[:notice] = 'アカウント名がすでに使われています'
        redirect '/register', 302
        return
      end

      query = 'INSERT INTO `users` (`account_name`, `passhash`) VALUES (?,?)'
      db.xquery(query,
        account_name,
        calculate_passhash(account_name, password)
      )

      session[:user] = {
        id: db.last_id
      }
      session[:csrf_token] = SecureRandom.hex(16)
      redirect '/', 302
    end

    get '/logout' do
      session.delete(:user)
      redirect '/', 302
    end

    get '/' do
      me = get_session_user()

      results = db.query('
        SELECT p.id, p.user_id, p.body, p.created_at, p.mime, u.account_name
        FROM `posts` AS p STRAIGHT_JOIN `users` AS u ON (p.user_id=u.id)
        WHERE u.del_flg=0
        ORDER BY p.created_at DESC
        LIMIT 20
      ')
      posts = make_posts(results)

      erb :index, layout: :layout, locals: { posts: posts, me: me }
    end

    get '/@:account_name' do
      user = db.xquery('SELECT * FROM `users` WHERE `account_name` = ? AND `del_flg` = 0',
        params[:account_name]
      ).first

      if user.nil?
        return 404
      end

      results = db.xquery('
        SELECT p.`id`, p.`user_id`, p.`body`, p.`mime`, p.`created_at`, u.`account_name`
        FROM `posts` p FORCE INDEX(`posts_user_idx`) JOIN`users` u ON (p.user_id=u.id)
        WHERE `user_id` = ? AND u.del_flg=0
        ORDER BY p.`created_at` DESC LIMIT 20
      ',
        user[:id]
      )
      posts = make_posts(results)

      comment_count = db.xquery('SELECT COUNT(*) AS count FROM `comments` WHERE `user_id` = ?',
        user[:id]
      ).first[:count]

      post_ids = db.xquery('SELECT `id` FROM `posts` WHERE `user_id` = ?',
        user[:id]
      ).map{|post| post[:id]}
      post_count = post_ids.length

      commented_count = 0
      if post_count > 0
        placeholder = (['?'] * post_ids.length).join(",")
        commented_count = db.xquery("SELECT COUNT(*) AS count FROM `comments` WHERE `post_id` IN (#{placeholder})",
          *post_ids
        ).first[:count]
      end

      me = get_session_user()

      erb :user, layout: :layout, locals: { posts: posts, user: user, post_count: post_count, comment_count: comment_count, commented_count: commented_count, me: me }
    end

    get '/posts' do
      max_created_at = params['max_created_at']
      results = db.xquery('
        SELECT p.`id`, p.`user_id`, p.`body`, p.`mime`, p.`created_at`, u.`account_name`
        FROM `posts` p STRAIGHT_JOIN `users` u ON (p.user_id=u.id)
        WHERE p.`created_at` <= ? AND u.del_flg=0
        ORDER BY p.`created_at` DESC
        LIMIT 20
      ',
        max_created_at.nil? ? nil : Time.iso8601(max_created_at).localtime
      )
      posts = make_posts(results)

      erb :posts, layout: false, locals: { posts: posts }
    end

    get '/posts/:id' do
      results = db.xquery('
        SELECT p.id, p.user_id, p.body, p.created_at, p.mime, u.account_name
        FROM `posts` p STRAIGHT_JOIN `users` u ON p.user_id=u.id
        WHERE p.`id` = ? AND u.del_flg=0
      ',
        params[:id]
      )
      posts = make_posts(results, all_comments: true)

      return 404 if posts.length == 0

      post = posts[0]

      me = get_session_user()

      erb :post, layout: :layout, locals: { post: post, me: me }
    end

    post '/' do
      me = get_session_user()

      if me.nil?
        redirect '/login', 302
      end

      if params['csrf_token'] != session[:csrf_token]
        return 422
      end

      if params['file']
        mime, ext = '', ''
        # 投稿のContent-Typeからファイルのタイプを決定する
        if params["file"][:type].include? "jpeg"
          mime, ext = "image/jpeg", "jpg"
        elsif params["file"][:type].include? "png"
          mime, ext = "image/png", "png"
        elsif params["file"][:type].include? "gif"
          mime, ext = "image/gif", "gif"
        else
          flash[:notice] = '投稿できる画像形式はjpgとpngとgifだけです'
          redirect '/', 302
        end

        if params['file'][:tempfile].size > UPLOAD_LIMIT
          flash[:notice] = 'ファイルサイズが大きすぎます'
          redirect '/', 302
        end

        query = 'INSERT INTO `posts` (`user_id`, `mime`, `imgdata`, `body`) VALUES (?,?,?,?)'
        db.xquery(query,
          me[:id],
          mime,
          '', # バイナリは保存しない
          params["body"],
        )
        pid = db.last_id

        # # アップロードされたテンポラリファイルをmvして配信ディレクトリに移動
        imgfile = IMAGE_DIR + "/#{pid}.#{ext}"
        FileUtils.mv(params['file'][:tempfile], imgfile)
        FileUtils.chmod(0644, imgfile)

        redirect "/posts/#{pid}", 302
      else
        flash[:notice] = '画像が必須です'
        redirect '/', 302
      end
    end

    get '/image/:id.:ext' do
      if params[:id].to_i == 0
        return ""
      end

      post = db.xquery('SELECT * FROM `posts` WHERE `id` = ?', params[:id].to_i).first

      if (params[:ext] == "jpg" && post[:mime] == "image/jpeg") ||
          (params[:ext] == "png" && post[:mime] == "image/png") ||
          (params[:ext] == "gif" && post[:mime] == "image/gif")
        headers['Content-Type'] = post[:mime]

        # 取得されたタイミングでファイルに書き出す
        imgfile = IMAGE_DIR + "/#{post[:id]}.#{params[:ext]}"
        f = File.open(imgfile, "w")
        f.write(post[:imgdata])
        f.close()
        return post[:imgdata]
      end

      return 404
    end

    post '/comment' do
      me = get_session_user()

      if me.nil?
        redirect '/login', 302
      end

      if params["csrf_token"] != session[:csrf_token]
        return 422
      end

      unless /\A[0-9]+\z/.match(params['post_id'])
        return 'post_idは整数のみです'
      end
      post_id = params['post_id']

      query = 'INSERT INTO `comments` (`post_id`, `user_id`, `comment`) VALUES (?,?,?)'
      db.xquery(query,
        post_id,
        me[:id],
        params['comment']
      )

      redirect "/posts/#{post_id}", 302
    end

    get '/admin/banned' do
      me = get_session_user()

      if me.nil?
        redirect '/login', 302
      end

      if me[:authority] == 0
        return 403
      end

      users = db.query('SELECT * FROM `users` WHERE `authority` = 0 AND `del_flg` = 0 ORDER BY `created_at` DESC')

      erb :banned, layout: :layout, locals: { users: users, me: me }
    end

    post '/admin/banned' do
      me = get_session_user()

      if me.nil?
        redirect '/', 302
      end

      if me[:authority] == 0
        return 403
      end

      if params['csrf_token'] != session[:csrf_token]
        return 422
      end

      query = 'UPDATE `users` SET `del_flg` = ? WHERE `id` = ?'

      params['uid'].each do |id|
        db.xquery(query, 1, id.to_i)
      end

      redirect '/admin/banned', 302
    end
  end
end
