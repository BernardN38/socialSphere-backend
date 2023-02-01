package com.socialsphere.notificationService.models;

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
    @CreationTimestamp
    private Timestamp timestamp;
    private String message;

    public Notification(String message) {
        this.message = message;
    }
}
