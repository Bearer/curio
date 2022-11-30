## URI

uri = URI("http://my.api.com/users/search?first_name=#{user.first_name}")
uri.query = URI.encode_www_form({ user: { first_name: "John", last_name: "Doe" } })

uri = URI('http://my.api.com/users/search')
user_1 = { first_name: "John", last_name: "Doe" }
uri.query = URI.encode_www_form(user_1)


## Net::HTTP

response = Net::HTTP.post_form(uri, { user: { first_name: "John", last_name: "Doe" } })


## Curl

User = Struct.new(:first_name, :last_name, keyword_init: true)
user_2 = User.new(first_name: "first", last_name: "last")
response = Curl.get("http://my.api.com/users/search?first_name=#{user_2.first_name}")

user_3 = { first_name: "John", last_name: "Doe" }
response = Curl.post("http://my.api.com/users/create", user_3)

response = Curl.post("http://my.api.com/users/create", { user: { first_name: "John", last_name: "Doe" } })


## RestClient

RestClient.post("http://my.api.com/users/create", { user: { first_name: "John", last_name: "Doe" } })


## Typhoeus

options = { body: { user: { first_name: "John", last_name: "Doe" } } }
response = Typhoeus.post("http://my.api.com/users/create", options)

response = Typhoeus.post("http://my.api.com/users/create", { body: { user: { first_name: "John", last_name: "Doe" } } })

Typhoeus.get("http://my.api.com/users/search?first_name=#{user_2.first_name}")


## HTTParty

HTTParty.get("http://my.api.com/users/search?first_name=#{user_2.first_name}")

params = {
	body: {
    user: {
      first_name: "John",
      last_name: "Doe",
    }
	}
}
HTTParty.post("http://my.api.com/users/create", params)

HTTParty.post("http://my.api.com/users/create", { body: { user: { first_name: "John", last_name: "Doe" } } })


## HTTP.rb

HTTP.get("http://my.api.com/users/search?first_name=#{user_2.first_name}")

HTTP.get("http://my.api.com/users/search", params: { user: { first_name: "John" } })

HTTP.post("http://my.api.com/users/create", form: { user: { first_name: "John", last_name: "Doe" } })


## Excon

Excon.get("http://my.api.com/users/search?first_name=#{user_2.first_name}")

Excon.post("http://my.api.com/users/create", body: { user: { first_name: "John", last_name: "Doe" } })


## Faraday

Faraday.get("http://my.api.com/users/search?first_name=#{user.first_name}")

params_2 = { user: { first_name: "John", last_name: "Doe" } }

encoded_params = URI.encode_www_form(params_2)

response = Faraday.post("http://my.api.com/users/create", encoded_params)

# Faraday.post("http://my.api.com/users/create") do |request|
#   request.body = { user: { first_name: "John", last_name: "Doe" } }.to_json
# end


## HTTPX

HTTPX.post("http://my.api.com/users/create", json: { user: { first_name: "John", last_name: "Doe" } })

HTTPX.get("http://my.api.com/users/search?first_name=#{user.first_name}")
