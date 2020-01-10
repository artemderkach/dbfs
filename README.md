# dbfs
http interface for storing files

## purpose
just for the memes

## starting service
- `docker-compose up --build` to start on `:8080` port
- `go build && ./dbfs` to start on `:8080` port 

## flow
- you should create your space for storing files  
`curl -X POST -d '{"email":"myEpicEmail"}' localhost:8080/register`  
you'll get token on email
- put this token in request header for accessing your workspace  
`curl -H "Authorization: MY_SECRET_TOKEN" ...`  
or you could put "Authorization: MY_SECRET_TOKEN" in file for request simplification  
`cull -H @path/to/file`  

## API
`POST /register` register user with given email  
`{ email: "42@mail.com" }`  
Next requests require "Authorization: TOKEN_VALUE" as header  
`GET /db` list root path  
`GET /db/some/path` list elements in given path  
`GET /db/some/file` download element (in case it is file)  
`POST /db/path/to file` write file (should be sent as data-binary request) to given path  
`DELETE /db/path/to/folder/or/file` deletes given element  

## environment variables

| environment    	| default value  |
|-----------------------|----------------|
| APP_PORT       	      | 8080           |
| DB_PATH             	| /tmp/mydb.bolt |
| MAILGUN_API_KEY      	|                |
| MAILGUN_ROOT_DOMAIN	|                |
| MAILGUN_SUBDOMAIN	   |                |

