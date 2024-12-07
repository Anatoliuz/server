# Project Name

A simple application that interacts with a PostgreSQL database using GORM and provides various endpoints to manage clients and commands.

---

## Table of Contents

- [Installation](#installation)
- [API Endpoints](#api-endpoints)
    - [Register Client](#register-client)
    - [List Clients](#list-clients)
    - [Execute Command](#execute-command)
    - [Save Command](#save-command)
    - [Get Next Command](#get-next-command)
    - [Complete Command](#complete-command)

---

## Installation

1. Clone the repository:
   ```
   bash
   git clone git@github.com:Anatoliuz/server.git
   cd server
2. Create a .env file in the project root with the following content:
   ```
      DB_USER=appuser
      DB_PASSWORD=apppassword
      DB_NAME=appdb
   ```
3. Build and start the application using Docker Compose:
   ```docker-compose up --build ```
   
4. Access the application at: 
   ```http://localhost:8080```

## API
### Register client
```
POST /register
     {
        "ip": "192.168.0.113",
        "port": "8080"
     }
```
###  Retrieves a list of all registered clients.
```
GET /clients
```
### Exec command on client
```
POST /client/{id}/e
     {
        "command": "reboot"
     }
```
### Add Command for client in queue
```
POST /client/{id}/command
{
  "command": "shutdown"
}
```
### Give all the remote client commands in queue (cmd should be sent by client)
```
GET /commands
[
  {
    "id": 1,
    "client_id": 1,
    "command": "shutdown",
    "status": "queued",
    "result": null
  },
  {
    "id": 2,
    "client_id": 1,
    "command": "reboot",
    "status": "completed",
    "result": "success"
  }
]
```
