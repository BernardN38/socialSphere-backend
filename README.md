# SocialSphere


![GitHub last commit](https://img.shields.io/github/last-commit/BernardN38/socialSphere-backend)
![GitHub top language](https://img.shields.io/github/languages/top/BernardN38/socialSphere-backend)
![Lines of code](https://img.shields.io/tokei/lines/github/BernardN38/socialSphere-backend)



This is the backend for a social media app called SocialSphere


# Table of contents
- [Description](#SocialSphere)
- [Architecture](#Architecture)
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


# Architecture
## Social Sphere Microservices Architecture
The Social Sphere app has adopted a microservices architecture to provide a modular and scalable approach to building the backend. This architecture enables the development and maintenance of each service independently, resulting in faster development and deployment times. In addition, each microservice can be scaled independently, allowing for better performance and availability during periods of high traffic.

## Benefits of Microservices Architecture
The microservices architecture used in Social Sphere also provides greater fault tolerance, as a failure in one microservice does not necessarily affect the others. The architecture also benefits from decoupling the microservices, which reduces latency to the client and provides a smoother and more responsive user experience.

The communication between the microservices is achieved using messaging with RabbitMQ. RabbitMQ allows the microservices to communicate with each other without being tightly coupled, which makes it easier to maintain and update the architecture over time.

# Microservices
## Media Microservice
The media microservice handles image and video uploads in the Social Sphere app. When a user uploads an image or video, the microservice receives the file and stores it in Minio, a blob storage service. If the file is larger than 5MB, the microservice sends a message on RabbitMQ to the image processing microservice, which compresses the image asynchronously to reduce its size. The media microservice also handles generating thumbnails and metadata for uploaded media.

## Post Service
The post service receives posts from clients in the Social Sphere app. These posts can include a body and an image. When a post is submitted, the post service stores it in a PostgreSQL database, along with metadata such as the author, date, and number of likes. The post service also handles retrieving posts from the database for display in the app.

## Authentication Service
The authentication service is responsible for handling user registration and login in the Social Sphere app. When a user registers, the service creates a new user account and stores the user's information in a PostgreSQL database. When a user logs in, the authentication service verifies the user's credentials and generates a JWT token, which is used to authenticate the user with other services in the app.

## Friend Service
The friend service handles friend requests and manages users' friend lists in the Social Sphere app. When a user sends a friend request, the friend service stores the request in a PostgreSQL database and sends a notification to the recipient. The friend service also handles retrieving users' friend lists and displaying them in the app.

## Messaging Service
The messaging service handles instant messaging in the Social Sphere app. When a user sends a message, the messaging service sends it over a WebSocket connection to the recipient. The messaging service also handles retrieving message history for display in the app.

## Identity Service
The identity service is responsible for storing user identifying information, such as their name, profile picture, and bio, in the Social Sphere app. When a user creates a profile, the identity service stores the user's information in a PostgreSQL database. The identity service also handles retrieving user information for display in the app.

## Image Processing Microservice
The image processing worker is a Python microservice that uses the PIL library to compress images asynchronously after it receives a message through RabbitMQ from the post service. It runs in the background and handles image compression without blocking the main thread of the app.

## Notification Service
The notification service sends real-time notifications to users in the Social Sphere app. This microservice is written in Java using the Spring Boot framework and utilizes a WebSocket connection to send notifications to clients. The service is responsible for sending notifications for new messages and new followers. When a user receives a new message or a new follower, the notification service sends a notification to the user's device via the WebSocket connection.



[(Back to top)](#table-of-contents)

# Author

The application was developed by a Bernardo Narvaez.

Bernardo Narvaez is a growth-oriented Full-stack Developer. Highly self-motivated. Skilled at problem solving and seeking multiple solutions to issues. Paying attention to details, while keeping an eye on long term goals.
[erisboxx@gmail.com](erisboxx@gmail.com)

[(Back to top)](#table-of-contents)
