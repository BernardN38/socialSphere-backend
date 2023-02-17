package com.socialsphere.notificationService.controller;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.socialsphere.notificationService.dto.FollowDto;
import com.socialsphere.notificationService.dto.MessageDto;
import com.socialsphere.notificationService.dto.NotificationDto;
import com.socialsphere.notificationService.models.Notification;
import com.socialsphere.notificationService.repository.NotificationRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.stream.Collectors;

@RestController
public class HttpController {
    @Autowired
    private NotificationRepository notificationRepository;
    @Autowired
    private SimpMessagingTemplate simpMessagingTemplate;

    @GetMapping("/notifications/{id}")
    public ResponseEntity<NotificationDto> getNotification(@PathVariable Long id) throws JsonProcessingException {
        Notification notification = notificationRepository.getReferenceById(id);
        ObjectMapper mapper = new ObjectMapper();
        MessageDto messageDto = mapper.readValue(notification.getPayload(), MessageDto.class);
        NotificationDto resp = new NotificationDto(
                notification.getUserId(),
                messageDto.toString(),
                notification.getType(),
                notification.getRead(),
                notification.getTimestamp()
        );
        System.out.println(resp);
        return ResponseEntity.ok(resp);
    }

    @GetMapping("/notifications")
    public List<NotificationDto> getNotifications() {
        List<Notification> notifications = notificationRepository.findAll();
        ObjectMapper mapper = new ObjectMapper();
        return notifications.stream()
                .map(notification -> {
                    return new NotificationDto(notification.getUserId(), notification.getPayload(), notification.getType(), notification.getRead(), notification.getTimestamp());

                })
                .collect(Collectors.toList());
    }

    @PostMapping("/notifications/follow")
    public ResponseEntity createNotification(@RequestBody FollowDto followDto) {
        System.out.println(followDto);
        Notification notification = new Notification();
        notification.setPayload(followDto.toString());
        notification.setType("newFollow");
        notification.setRead(false);
        notification.setUserId(Long.valueOf(followDto.getFollowed()));
        Notification resp = notificationRepository.save(notification);
        simpMessagingTemplate.convertAndSendToUser(String.valueOf(followDto.getFollowed()), "/notifications", notification);
        return ResponseEntity.ok(resp);
    }

}
