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

* Clone this repository to your computer

    `git clone https://github.com/ValeriiaGrebneva/Chirpy`

* Get your connection string
    
    Your connection string is a URL with the format:

    `protocol://username:password@host:port/database`

    In my case, it was `postgres://postgres:postgres@localhost:5432/chirpy`

* Install _Goose_ and do migrations

    https://github.com/pressly/goose#install
    
    Migrations can be done using (do not forget to change to _your_ connection string): 
    
    `goose -dir sql/schema postgres postgres://postgres:postgres@localhost:5432/chirpy up`

* create .env file in the root of the project with the following information:

    ```
    DB_URL="YOUR_CONNECTION_STRING_HERE"
    PLATFORM="dev"
    KEY_JWT="JWT_KEY_HERE"
    POLKA_KEY="POLKA_KEY_HERE"
    ```

    where DB_URL is a database connection string, PLATFORM string is for accessing _.../admin/..._ endpoints, KEY_JWT is a secret string for JWTs, and POLKA_KEY is used to verify webhook.

* Build and run the server

    `go build -o out && ./out`


### Endpoints:

The website will be available through the link _http://localhost:8080/app/_

The available endpoints (_http://localhost:8080/..._):

1. GET _.../api/healthz_ - returns 200 status code if the server is running;

2. GET _.../admin/metrics_ - returns a text with a number of how many times the webpage _http://localhost:8080/app/_ was visited;

3. POST _.../admin/reset_ - resets the visiting number for _.../api/metrics_ endpoint to zero and deletes all the users;

4. POST _.../api/chirps_ - accepts a JSON file with:

    ```
    "body": "Here is the text of the Chirp",
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
    ```

    It also required to have a valid JWT (JSON Web Token). This places the chirp to database, and returns a JSON file with the chirp's information;

5. GET _.../api/chirps_ - returns all the chirps from the database as an array sorted by creation date in ascending order with optional parameters:
    * author_id (_.../api/chirps?author_id=1_) - endpoint will return only the chirps for that author, otherwise return all chirps,
    * desc order (_...api/chirps?sort=desc_) - endpoint will sort in descending order instead;

6. GET _.../api/chirps/{chirpID}_ - returns the chirp with this ID;

7. DELETE _.../api/chirps/{chirpID}_ - deletes the certain chirp if the current user (checking through the token in the header) is the author of the chirp;

8. POST _.../api/users_ - creates a user with accepted email and password:

    ```
    "password": "password",
    "email": "user@example.com"
    ```

    and returns user's data;

9. POST _.../api/login_ - accepts a JSON file with password, email:

    ```
    "password": "password",
    "email": "user@example.com"
    ```

    This allows user to login if the password is validated, then returns user's information, as well as access (JWT, valid for 1 hour) and refresh tokens (valid for 60 days);

10. POST _.../api/refresh_ - requires a refresh token in the header `Authorization: Bearer <token>` and checks if it is valid. If yes, it returns an access token (JWT);

11. POST _.../api/revoke_ - requires a refresh token in the header `Authorization: Bearer <token>` and revokes the token if it is valid;

12. PUT _.../api/users_ - requires an access token in the header and a new password and email in the request:

    ```
    "password": "password",
    "email": "user@example.com"
    ```

    If successful, returns 200 status code and user's information;

13. POST _.../api/polka/webhooks_ - test endpoint for external imaginary service that is supposed to give information if a user has a subscription. The endpoint accepts:

    ```
    {
        "event": "user.upgraded",
        "data": {
            "user_id": "3311741c-680c-4546-99f3-fc9efac2036c"
        }
    }
    ```

    The request supposed to have the API key that matches the one in your .env file. If the event is "user.upgraded" and there is such user in database, the user is marked as a Chirpy Red member in the database.


##

While doing this Chirpy project, I learned how to:
* build an HTTP server in Go
* handle communication client-server with RESTful APIs using headers, JSON, and status codes
* use PostgreSQL database, create migrations and queries for keeping and retrieving data (Goose and SQLC are helpful tools)
* create authentication and authorization systems
* implement webhooks