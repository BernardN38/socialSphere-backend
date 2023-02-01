package com.socialsphere.notificationService.controller;

import com.socialsphere.notificationService.dto.NotificationDto;
import com.socialsphere.notificationService.models.Notification;
import com.socialsphere.notificationService.repository.NotificationRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.sql.Timestamp;

@RestController
public class Controller {
    @Autowired
    private NotificationRepository notificationRepository;
    @GetMapping("/notifications/{id}")
    public NotificationDto getNotification(@PathVariable Long id){
        Notification notification = notificationRepository.getReferenceById(id);
        NotificationDto resp = new NotificationDto(
            notification.getId(),
                notification.getTimestamp(),
                notification.getMessage()
        );
        return resp;
    }

    @PostMapping("/notifications")
    public void createNotification(@RequestBody NotificationDto notificationDto){
        Notification notification = new Notification(
                notificationDto.getMessage()
        );
        notificationRepository.save(notification);
    }
}
