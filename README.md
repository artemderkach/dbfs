# dbfs
file system in database

## starting service
- `docker-compose up --build` to start on `:8080` port  
- `go build && ./dbfs` to start on `:8080` port 


## purpose
just for the memes

## api
- `GET /view` return the current state of database
- `POST /put` create record with file in database  
   expect "multipart/form-data" format.
