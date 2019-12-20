# dbfs
http interface for storing files

## purpose
just for the memes

## starting service
- `docker-compose up --build` to start on `:8080` port
- `go build && ./dbfs` to start on `:8080` port 

## flow
first you should create your space for storing files  
`curl -X -d '{"email":"myEpicEmail"}' localhost:8080/register`  
you'll get token on email

## environment variables

| environment    | default value  |   |   |   |
|----------------|----------------|---|---|---|
| APP_PORT       | 8080           |   |   |   |
| DB_PATH        | /tmp/mydb.bolt |   |   |   |
| MAIGUN_API_KEY |                |   |   |   |
| MAINLUN_DOMAIN |                |   |   |   |

## api
`{collection}` could be one of `private` or `public`  

for `private` route, `Custom-Auth` header should be set. Value should be your password  

- `GET /{collection}/view` return the current state of database
- `POST /{collection}/{this/will/be/my/filepath}` create record with file in database  
   expect "multipart/form-data" format.
- `GET /{collection}/download/{this/will/be/my/filepath/andFilename}` return content of "filename" 
- `DELETE /{collection}/{this/will/be/my/filepathOrFilename}` returns `GET /view` after deletion

## examples
```
$ curl -F file=@.bashrc localhost:8080/public/put
.bashrc
$
$ curl -F file=@.emacs localhost:8080/public/put
.bashrc
.emacs
$ curl -X "DELETE" localhost:8080/public/.bashrc
.emacs

$ curl -F file=@.bashrc localhost:8080/private/put
permission denied
$ curl -F file=@.bashrc -H "Custom-Auth: mypass" localhost:8080/private/put
.bashrc
$ curl -H "Custom-Auth: my_pass" localhost:8080/private/download/.bashrc > .my_bashrc
```
