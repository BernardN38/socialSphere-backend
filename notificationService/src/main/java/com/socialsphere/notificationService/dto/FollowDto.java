package com.socialsphere.notificationService.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.gson.Gson;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
public class FollowDto {
    private int follower;
    private String followerUsername;
    private  int followed;

    private String messageType;

    @Override
    public String toString() {
        String json = new Gson().toJson(this);
        return  json;
    }
}