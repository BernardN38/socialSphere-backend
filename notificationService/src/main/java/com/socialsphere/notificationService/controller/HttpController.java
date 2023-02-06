package com.socialsphere.notificationService.controller;

import com.socialsphere.notificationService.dto.NotificationDto;
import com.socialsphere.notificationService.models.Notification;
import com.socialsphere.notificationService.repository.NotificationRepository;
import org.springframework.beans.factory.annotation.Autowired;
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
//    @GetMapping("/notifications/{id}")
//    public NotificationDto getNotification(@PathVariable Long id){
//        Notification notification = notificationRepository.getReferenceById(id);
//        NotificationDto resp = new NotificationDto(
//            notification.getUserId(),
//                notification.getFromUsername(),
//                notification.getPayload(),
//                notification.getType(),
//                notification.getTimestamp()
//        );
//        return resp;
//    }
//    @GetMapping("/notifications")
//    public List<NotificationDto> getNotifications(){
//        List<Notification> notifications = notificationRepository.findAll();
//        return notifications.stream()
//                .map(notification -> new NotificationDto(notification.getUserId(), notification.getFromUsername(), notification.getPayload(), notification.getType(),notification.getTimestamp()))
//                .collect(Collectors.toList());
//    }
//    @PostMapping("/notifications")
//    public void createNotification(@RequestBody NotificationDto notificationDto){
//        Notification notification = new Notification(
//                notificationDto.getPayload()
//        );
//        notificationRepository.save(notification);
//    }


    @PostMapping("/messagews")
    public String sendMessage(@RequestBody NotificationDto notification) {
        simpMessagingTemplate.convertAndSendToUser(notification.getUserId().toString(),"/notifications", notification);
        return "test";
    }

}
