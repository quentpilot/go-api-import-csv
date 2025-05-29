# Import and read a CSV file using Gin, RabbitMQ and Goroutines
This project is a sandbox to test the power of Go.

I wish I could compare performances between a Go and a PHP application... but the lack of PHP threads handler leaves me in doubt.

## ⚙️ Overall of the process
1. If the file is a valid CSV
    1. The file is saved through a `/shared` volume
    2. Then, an `AMQP` message is published to be consumed soon
2. The worker get the message and analysis the file
    1. If the number of rows is less than 25k (by default), then it read the whole file
    2. But, if the number of rows is more than 25k
        1. then it chunks the file as many time it need to save them to `/tmp` directory
        2. and read each file within a `goroutine`
3. Finally, all files are deleted

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
