# Insecure FTP

class User
  attr_reader :name, :email, :password
end

## Detected
require "net/ftp"

ftp = Net::FTP.new("ftp.ruby-lang.org")
ftp.login("anonymous", "matz@ruby-lang.org")
ftp.chdir("/pub/ruby")
tgz = ftp.list("ruby-*.tar.gz").sort.last
ftp.getbinaryfile(tgz, tgz)
ftp.close

Net::FTP.open('example.com') do |ftp|
  ftp.login
  files = ftp.chdir('pub/lang/ruby/contrib')
  files = ftp.list('n*')
  ftp.getbinaryfile('nif.rb-0.91.gz', 'nif.gz', 1024)
end

Net::FTP.open("ftp.site.com") do |ftp|
  file = Tempfile.new("user_data")
  begin
    file << { user: { ethnicity: "martian" } }.to_json
    file.close

    ftp.puttextfile(file.path, "/users/123.json")
  ensure
    file.close!
  end
end

## Not detected
require "net/sftp"

Net::SFTP.start("localhost", "user") do |sftp|
  sftp.upload! "/local/file.tgz", "/remote/file.tgz"
end
