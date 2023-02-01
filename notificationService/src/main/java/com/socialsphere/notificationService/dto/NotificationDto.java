package com.socialsphere.notificationService.dto;

import lombok.AllArgsConstructor;
import lombok.Data;

import java.sql.Timestamp;

@AllArgsConstructor
@Data
public class NotificationDto {
    private Long id;
    private Timestamp timestamp;
    private String message;
}
