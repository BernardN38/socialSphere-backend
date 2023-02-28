package com.socialsphere.notificationService.models;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.gson.Gson;
import com.socialsphere.notificationService.dto.MessageDto;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;
import org.hibernate.annotations.CreationTimestamp;

import java.sql.Timestamp;
@NoArgsConstructor
@AllArgsConstructor
@Setter
@Getter
@Entity
public class Notification {
    @Id
    @GeneratedValue(strategy=GenerationType.AUTO)
    private Long id;
    private Long userId;
    private String payload;
    private String type;
    private Boolean read = false;
    @CreationTimestamp
    private Timestamp createdAt;

    public Notification(Long userId, String payload,  String type) {
        this.userId = userId;
        this.payload = payload;
        this.type = type;
    }


}
