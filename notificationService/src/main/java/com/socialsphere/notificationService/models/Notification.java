package com.socialsphere.notificationService.models;

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

    private String fromUsername;
    private String type;
    @CreationTimestamp
    private Timestamp timestamp;
    public Notification(String payload) {
        this.payload = payload;
    }
}
