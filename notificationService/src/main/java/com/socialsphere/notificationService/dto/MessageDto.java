package com.socialsphere.notificationService.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.json.JsonMapper;
import com.google.gson.Gson;
import jakarta.persistence.*;
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
    @JsonProperty
    private int fromUserId;
    @JsonProperty
    private String fromUsername;
    @JsonProperty
    private int toUserId;
    @JsonProperty
    private String  subject;
    @JsonProperty
    private String message;

    public MessageDto(String json) throws JsonProcessingException {
        ObjectMapper mapper = new ObjectMapper();
        MessageDto messageDto = mapper.readValue(json, MessageDto.class);
        this.fromUserId = messageDto.getFromUserId();
        this.fromUsername = messageDto.getFromUsername();
        this.toUserId = messageDto.getToUserId();
        this.subject = messageDto.getSubject();
        this.message = messageDto.getMessage();
    }
    @Override
    public String toString() {
        String json = new Gson().toJson(this);
        return  json;
    }
}