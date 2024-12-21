Required Libraries:

Go and Postgres

Compilation:

go install gator

Commands:

	register	registers a user.  requires the user name as a parameter
	reset		resets the database
	users		prints a list of users
	agg		gathers all registered RSS feeds, scan rate is passed in as a parameter
	addfeed		adds a feed, requires the name of the feed and a source URL
	feeds		displays a list of feeds
	follow		adds a feed to the logged in user.  requires the URL as a parameter
	following	displays a list of feeds followed by the logged in user.  
	unfollow	removes a followed URL for the logged in user.  requires an URL as a parameter
	browse		displays the most recent feeds for the logged in user.  Option limit parameter (default 2)

The user config file is located in the user's home directory and is called .gatorconfig.json.  The format of this file is as follows:

{
	"db_url":{connection string},
	"current_user_name":{optional username}
}
