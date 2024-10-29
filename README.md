## ![alt text](./asssets/logo.svg) smolDB

A JSON document-based database that relies on key based access to achieve O(1) access time.

`smolDB` is a document database with key-based access and reference resolution, all documents are on-disk, human-readable, and can be accessed through a REST API, making them very easy for debugging.

However, the database does not support any advanced queries, sharding, or support for storage distribution.

### key principles
- Quick access — key-based retrieval in O(1) time
- Simple to troubleshoot — all documents are stored as human-readable JSON files
- Easy to deploy — a single executable with no external dependencies  no need for language-specific drivers!

### endpoints
#### `GET /`
```bash
# check the heath of the database
curl localhost:8080/

# example output on 200 OK
# > {"message":["smolDB is working fine!"]}
```

#### `GET /keys`
```bash
# get all files in database index
curl localhost:8080/keys

# example output on 200 OK
# > {"files":["test","test2","test3"]}
```

#### `POST /regenerate`
```bash
# manually regenerate index
# shouldn't need to be done as each operation should update index on its own
curl -X POST localhost:8080/regenerate

# example output on 200 OK
# > regenerated index
```

#### `GET /key/:key`
```bash
# get document with key `test`
curl localhost:8080/key/test

# example output on 200 OK (found key)
# > {"example_field": "example_value"}
# example output on 404 NotFound (key not found)
# > key 'test' not found
```

#### `PUT /key/:key`
```bash
# creates document `test` if it doesn't exist
# otherwise, replaces content of `key` with given
curl -X PUT -H "Content-Type: application/json" \
            -d '{"key1":"value"}' localhost:8080/key/test

# example output on 200 OK (create/update success)
# > create 'test' successful
```

#### `DELETE /key/:key`
```bash
# deletes document `test`
curl -X DELETE localhost:8080/key/test

# example output on 200 OK (delete success)
# > delete 'test' successful
# example output on 404 NotFound (key not found)
# > key 'test' doest not exist
```

#### `GET /key/:key/field/:field`
```bash
# get `example_field` of document `test`
curl localhost:8080/key/test/field/example_field

# example output on 200 OK (found field)
# > "example_value"
# example output on 400 BadRequest (field not found)
# > err key 'test' does not have field 'example_field'
# example output on 404 NotFound (key not found)
# > key 'test' not found
```
#### `PATCH /key/:key/field/:field`
```bash
# update `field` of document `test` with content
# if field doesnt exist, create it
curl -X PATCH -H "Content-Type: application/json" \
              -d '{"nested":"json!"}' \
              localhost:8080/key/test/field/example_field

# example output on 200 OK (found field)
# > patch field 'example_field' of key 'key' successful
# example output on 404 NotFound (key not found)
# > key 'test' not found
```

### building `smolDB` from scratch
- Run `git clone https://github.com/themillenniumfalcon/smolDB`
- Run `make build`
- For cross-platform builds, run `make build-all` (optional)