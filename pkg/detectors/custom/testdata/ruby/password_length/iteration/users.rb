Devise.setup do |config|
	config.password_length = 6
end

# it should ignore this one as value is above minimum
Devise.setup do |config|
	config.password_length = 11
end