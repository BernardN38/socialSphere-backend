package com.socialsphere.notificationService.dto;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.json.JsonMapper;
import com.google.gson.Gson;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.sql.Timestamp;

@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
public class MessageDto {
    private int fromUserId;
    private String fromUsername;
    private int toUserId;
    private String  subject;
    private String message;
    private Timestamp timestamp;

    public MessageDto(String json) throws JsonProcessingException {
        ObjectMapper mapper = new ObjectMapper();
        MessageDto messageDto = mapper.readValue(json, MessageDto.class);
        this.fromUserId = messageDto.getFromUserId();
        this.fromUsername = messageDto.getFromUsername();
        this.toUserId = messageDto.getToUserId();
        this.subject = messageDto.getSubject();
        this.message = messageDto.getMessage();
        this.timestamp = messageDto.getTimestamp();
    }
    @Override
    public String toString() {
        String json = new Gson().toJson(this);
        return  json;
    }
}