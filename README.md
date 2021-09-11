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
`curl -H @path/to/file`  

## API
`POST /register` register user with given email  
`{ email: "42@mail.com" }`  
Next requests require "Authorization: TOKEN_VALUE" as header  
`GET /db` list root path  
`POST /db` write file (should be sent as data-binary request) to given path  
`DELETE /db` deletes given element  
`GET /share` copies node to publick space  
`GET /shared` get shared data  
`GET /help` API routes  
`GET /examples` return requests examples  

## environment variables

| environment    	| default value  |
|-----------------------|----------------|
| APP_PORT       	      | 8080           |
| DB_PATH             	| /tmp/mydb.bolt |
| MAILGUN_API_KEY      	|                |
| MAILGUN_ROOT_DOMAIN	|                |
| MAILGUN_SUBDOMAIN	   |                |

## examples
`curl -w '\n' -X POST -d '{"email": "myEpicEmail@gmail.com"}' localhost:8080/register` register with given email  
`curl -w '\n' -X POST -H @$HOME/Documents/dbfs_headers -d $HOME/data.txt localhost:8080/db/data.txt` write file with (use auth headers from file)  
`curl -w '\n' -X POST -H "Authorization: <toke>" -d $HOME/data.txt localhost:8080/db/data.txt` write file with (use auth direct)  

`curl -w '\n' -X GET -H @$HOME/Documents/dbfs_headers localhost:8080/db/data.txt` dowload written file  
`curl -w '\n' -X GET -H @$HOME/Documents/dbfs_headers localhost:8080/db` view root tree  

`curl -w '\n' -X DELETE -H @$HOME/Documents/dbfs_headers localhost:8080/db/data.txt` delete data file  
`curl -w '\n' -X DELETE -H @$HOME/Documents/dbfs_headers localhost:8080/db/someFolder` in case of folder, will delete it and all it's childs  

`curl -w '\n' -X GET -H @$HOME/Documents/dbfs_headers localhost:8080/share/someFolder` share folder (after which, you can access to it without auth header. you can find token it root tree  
`curl -w '\n' -X GET -H @$HOME/Documents/dbfs_headers localhost:8080/shared/<token>/someFolder` view or download shared data  
