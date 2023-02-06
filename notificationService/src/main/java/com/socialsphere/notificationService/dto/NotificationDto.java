package com.socialsphere.notificationService.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.*;

import java.sql.Timestamp;

@NoArgsConstructor
@AllArgsConstructor
@Getter
@Setter
public class NotificationDto {
    private Long userId;
    private String fromUsername;
    private MessageDto payload;
    private String type;
    private Timestamp timestamp = new Timestamp(System.currentTimeMillis());

    public NotificationDto(String message) throws JsonProcessingException {
        ObjectMapper mapper = new ObjectMapper();
        MessageDto messageDto = mapper.readValue(message, MessageDto.class);
        this.payload = messageDto;
    }
}
