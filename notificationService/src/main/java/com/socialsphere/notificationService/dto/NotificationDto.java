package com.socialsphere.notificationService.dto;

import lombok.*;

import java.sql.Timestamp;

@NoArgsConstructor
@AllArgsConstructor
@Getter
@Setter
public class NotificationDto {
    private Long userId;
    private String message;
    private String type;
    private Timestamp timestamp = new Timestamp(System.currentTimeMillis());

    public NotificationDto(String message) {
        this.message = message;
    }
}
