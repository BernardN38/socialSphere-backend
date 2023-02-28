# SocialSphere


![GitHub last commit](https://img.shields.io/github/last-commit/BernardN38/socialSphere-backend)
![GitHub top language](https://img.shields.io/github/languages/top/BernardN38/socialSphere-backend)
![Lines of code](https://img.shields.io/tokei/lines/github/BernardN38/socialSphere-backend)



This is the backend for a social media app called SocialSphere


# Table of contents
- [Description](#SocialSphere)
- [Table of contents](#table-of-contents)
- [Installation](#installation)
- [Development](#development)
- [Author](#author)

# Installation

This app requires the installation of docker-compose, make,and Docker.


The following set of instructions will allow you to clone the application and run it locally on your device. Note that all of the following lines of code are ran in the terminal.

1. From within the terminal, `cd` into the directory of your choice and run the following command:

        ```
        git clone https://github.com/BernardN38/socialSphere-backend
         ```

2. `cd` into the the application: 
	
	```
	cd socialSphere-backend/project
	```

3. Install the dependecies and run application:

	```
	make up_build
	```

	
5. Note that the app will open up on localhost 8080 (http://localhost:8080/). 
6. Swagger documentation can be seen at http://localhost:8080/swagger-ui.html

Notes.


	
[(Back to top)](#table-of-contents)



# Development




<img src="https://raw.githubusercontent.com/devicons/devicon/1119b9f84c0290e0f0b38982099a2bd027a48bf1/icons/java/java-original.svg" alt="Java Logo" height="50px" width="50px"><img src="https://raw.githubusercontent.com/devicons/devicon/1119b9f84c0290e0f0b38982099a2bd027a48bf1/icons/spring/spring-original.svg" alt="Spring Logo" height="50px" width="50px">

Here are the languages and technologies used in the project, along with brief descriptions:

- Go:
  - Primary language used for most of the microservices.
  - Go-Chi:
    - A small, lightweight library used to route requests.
  - Sqlc:
    - A code generator that converts SQL queries into type-safe Go code, used to interact with the PostgreSQL database.
- Python:
  - PIL:
    - An image processing library used to compress and optimize images.
- Java:
  - Spring Boot:
    - A web framework used to build the notification microservice HTTP server and WebSocket.
- PostgreSQL:
  - A SQL relational database used to provide persistence to the backend.
- MongoDB:
  - A NoSQL database used to provide persistence to the messaging microservice.
- NGINX:
  - A web server used as a reverse proxy to make a unified front for all microservices.
- Docker:
  - A container platform used to run the microservices and containers.
- Minio:
  - An S3-like blob storage used to store images and videos.
- RabbitMQ:
  - A messaging queue used for communication between microservices to decouple.
- JWT:
  - Used to secure and provide authentication to all microservices.






[(Back to top)](#table-of-contents)

# Author

The application was developed by a Bernardo Narvaez.

Bernardo Narvaez is a growth-oriented Full-stack Developer. Highly self-motivated. Skilled at problem solving and seeking multiple solutions to issues. Paying attention to details, while keeping an eye on long term goals.
[erisboxx@gmail.com](erisboxx@gmail.com)

[(Back to top)](#table-of-contents)
