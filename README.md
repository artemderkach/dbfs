# dbfs
file system in database

## purpose
just for the memes

## starting service
- `docker-compose up --build` to start on `:8080` port
- `go build && ./dbfs` to start on `:8080` port 

## api
`{collection}` could be one of `private` or `public`  

for `public` route, `Custom-Auth` header should be set. Value should be your password  

- `GET /{collection}/view` return the current state of database
- `POST /{collection}/put` create record with file in database  
   expect "multipart/form-data" format.
- `GET /{collection}/download/{filename}` return content of "filename" 
- `DELETE /{collection}/delete/{filename}` returns `GET /view` after deletion
