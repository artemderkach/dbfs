# dbfs
file system in database

## purpose
just for the memes

## starting service
- `docker-compose up --build` to start on `:8080` port  
- `go build && ./dbfs` to start on `:8080` port 

## api
- `GET /view` return the current state of database
- `POST /put` create record with file in database  
   expect "multipart/form-data" format.
- `GET /download/{filename}` return content of "filename" 
- `DELETE /delete/{filename}` returns `GET /view` after deletion
