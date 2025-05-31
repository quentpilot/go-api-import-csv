# Import and read a CSV file using Gin, RabbitMQ and Goroutines
This project is a sandbox to test the power of Go.

I wish I could compare performances between a Go and a PHP application... but the lack of PHP threads handler leaves me in doubt.

## ⚙️ Overall of the process
1. If the file is a valid CSV
    1. The file is saved through a `/shared` volume
    2. Then, an `AMQP` message is published to be consumed soon
2. The worker get the message and analysis the file
    1. If the number of rows is less than `6k` (by default), then it read the whole file
    2. But, if the number of rows is more than 6k
        1. then it chunks the file as many time it need to save them to `/tmp` directory
        2. and read each file within a `goroutine`
    3. Insert batch of `3k` rows (by default) through `MySQL` database
3. Finally, all files are deleted, AMQP message is acknowleged
 

## ✅ Tests
Run command `make test` to run all working tests

## 🚀 Setup
### 🛠️ Makefile
Run command `make run` to build docker and run all services.

### 🐳 Docker
Run command `docker-compose up --build`

## 📂 Logs
You can follow the file consuming process from `STDOUT` or from files `logs/api.log` and `logs/worker.log`

## 🧪 Try API
After running project, call endpoint POST [localhost:8080/upload](http://localhost:8080/upload) with `file` parameter.

You can test with your favorite Postman, Insomnia, or even cURL.

### ⚡ Quick cURL examples
CSV file with less than 10 rows
``` 
curl --location 'http://localhost:8080/upload' --form 'file=@testdata/contacts_light.csv'
```

CSV file with 100 000 rows
``` 
curl --location 'http://localhost:8080/upload' --form 'file=@testdata/contacts_100k.csv'
```

### 🧠 Create your own CSV
By using command `make generate-csv CSV_LINES=1_000_000`,

this will create file `testdata/gen_contact_1000000.csv` with 1M of rows.

Then, you can use it like above cURL examples.

## 📕 API Doc
### HTTP server health status

<details>
 <summary><code>GET</code> <code><b>/ping</b></code> <code>(Checks that server is running)</code></summary>

#### Parameters

> | name      |  type                | content-type            | description                          
> |-----------|----------------------|-------------------------|--------------------------------------
> | none      |  none                | none                    | none 


#### Responses

> | http code     | content-type                      | response                                       |
> |---------------|-----------------------------------|------------------------------------------------|
> | `200`         | `application/json`                | `{"message": "API is running", "status": "ok"}`|
> | `500`         | `text/html`                       | none                                           |

#### Example cURL

> ```bash
>  curl --location 'localhost:8080/ping'
> ```

</details>

### Upload Contacts

<details>
 <summary><code>POST</code> <code><b>/upload</b></code> <code>(Stores file to process it from AMQP queue)</code></summary>

#### Parameters

> | name      |  type                | content-type            | description                          |  expected csv headers format     |
> |-----------|----------------------|-------------------------|--------------------------------------|----------------------------------|
> | file      |  multipart/form-data | text/csv                |  CSV file containing customers infos | "Phone";"Firstname";"Lastname"   |


#### Responses

> | http code     | content-type                      | response                                                                                                                |
> |---------------|-----------------------------------|-------------------------------------------------------------------------------------------------------------------------|
> | `202`         | `application/json`                | `{"message": "File is being processed", "status_url": "http://localhost:8080/upload/status/{uuid}", "uuid": "{uuid}"}`  |
> | `400`         | `application/json`                | `{"message":"Missing File"}`                                                                                            |
> | `415`         | `application/json`                | `{"message":"invalid file type {ext}. expected a .csv file"}`                                                           |
> | `500`         | `application/json`                | `{"message":"Cannot save file"}`                                                                                        |

##### Success
```json 
{
    "message": "File is being processed", 
    "status_url": "http://localhost:8080/upload/status/{uuid}", // Callback URL to follow file upload progress
    "uuid": "{uuid}"                                            // Uuid of the request to handle contacts
}
```

#### Example cURL

> ```bash
>  curl --location 'http://localhost:8080/upload' --form 'file=@testdata/contacts_light.csv'
> ```

</details>

### Upload File Status

<details>
 <summary><code>GET</code> <code><b>/upload/status/{uuid}</b></code> <code>(Checks in real-time file processing status)</code></summary>

#### Parameters

> | name      |  type                | content-type            | description                            |
> |-----------|----------------------|-------------------------|----------------------------------------|
> | uuid      |  string              | text/html               |  identifier of file linked to contacts |


#### Responses

> | http code     | content-type                      | response                                                                                         |
> |---------------|-----------------------------------|--------------------------------------------------------------------------------------------------|
> | `200`         | `application/json`                | `{"Status": "Scheduled/Processing/Completed", "Total": 10, "Inserted": 8, "Percentile": 80.000}` |
> | `404`         | `application/json`                | `{"message":"Progress Status Not Found"}`                                                        |
> | `504`         | `application/json`                | `{"message":"Request to worker timed out"}`                                                      |
> | `500`         | `application/json`                | `{"message":"Failed to get progress status from worker"}`                                        |
> | `502`         | `application/json`                | `{"message":"Corrupted progress status data"}`                                                   |

##### Success
```json 
{
    "Status": "Scheduled/Processing/Completed", // Humanized process status
    "Total": 10,                                // Total file rows (subtitute CSV headers)
    "Inserted": 8,                              // Total inserted rows through database
    "Percentile": 80.000                        // Progress Percentile
}
```

#### Example cURL

> ```bash
>  curl --location 'http://localhost:8080/upload/status/7b1cdab9-40eb-49a3-bced-7523b8a3590e'
> ```

</details>

## ⚙️ Environment
A `.env` file allows to set default app configuration.

### 🔥 Hot Reload
A `goroutine` listen to a `SIGHUP` signal to reload `.env` configuration like logs level.

When loading Docker containers, command `make reload` sends a signal to hot reload app configuration.

### 💡 List of useful variables

| Name                    | Default Value | Hot Reload          | Description
| :---------------------- | :-----------  | :-----------------: | :-------------------
| LOG_LEVEL               |  INFO         |         ✅          | Display log level
| AMQP_DSN                |  import_queue |         ❌          | AMQP server auth
| AMQP_QUEUE              |  INFO         |         ❌          | AMQP queue name
| AMQP_LIFETIME           |  60           |         ✅          | AMPQ message timeout in seconds
| HTTP_PORT               |  INFO         |         ❌          | Web API port
| HTTP_MAX_CONTENT_LENTGH | 10485760      |         ❌          | Max API request size
| FILE_CHUNK_LIMIT        | 6000          |         ✅          | Max rows by file (auto chunked if reached)
| FILE_UPLOAD_TIMEOUT     | 30            |         ✅          | Timeout in seconds for each chunked file to upload
| BATCH_INSERT            | 3000          |         ✅          | Number of rows by SQL INSERT


## 🕙 Roadmap
1. Implement testing
2. Benchmark performances with time and memory consumption
    1. Tests with empty datatable to insert 1M rows
    2. Tests with already filled datatable of 1M rows or more
3. Add PHP/Symfony Docker service to create the same API/Test/Benchmark scenario
4. Add a REST response to /upload with another endpoint (uuid) to follow upload status in percentile
5. Implements a GUI to select a file to upload and an a progressbar
