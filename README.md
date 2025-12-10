**_This README file is still in progress_**

### What you will need for this application to work:

* Use a command line

* Install _Go_ (version 1.22+ is required)

    https://webinstall.dev/golang/

* Install _PostgreSQL_ and set up the database

    https://webinstall.dev/postgres/
    
        sudo -u postgres psql
        CREATE DATABASE gator;
        \c gator
        ALTER USER postgres PASSWORD 'postgres';

    You can type exit to leave the psql shell.

* Download the files from this repository to your computer

    `git clone https://github.com/ValeriiaGrebneva/Chirpy`

* Get your connection string
    
    Your connection string is a URL with the format:

    `protocol://username:password@host:port/database`

    In my case, it was `postgres://postgres:postgres@localhost:5432/chirpy`

* Install _Goose_ and do migrations

    https://github.com/pressly/goose#install
    
    Migrations can be done using (do not forget to change to _your_ connection string): 
    
    `goose -dir sql/schema postgres postgres://postgres:postgres@localhost:5432/chirpy up`

* Build and run the server

    go build -o out && ./out



##

While doing this Chirpy project, I learned how to:
* build an HTTP server in Go
* handle communication client-server with RESTful APIs using headers, JSON, and status codes
* use PostgreSQL database, create migrations and queries for keeping and retrieving data (Goose and SQLC are helpful tools)
* create authentication and authorization systems
* implement webhooks