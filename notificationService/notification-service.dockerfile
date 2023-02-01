# Use an official Java runtime as the base image
FROM openjdk:19

# Set the working directory in the container to /app
WORKDIR /app

# Copy the JAR file of the Spring Boot application to the container
COPY target/*.jar app.jar

# Expose port 8080 to allow access to the application
EXPOSE 8080

# Set the command to run the Spring Boot application
CMD ["java", "-jar", "app.jar"]