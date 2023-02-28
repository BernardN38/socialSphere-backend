package com.socialsphere.notificationService.controller;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.socialsphere.notificationService.dto.FollowDto;
import com.socialsphere.notificationService.dto.MessageDto;
import com.socialsphere.notificationService.dto.NotificationDto;
import com.socialsphere.notificationService.dto.UserDto;
import com.socialsphere.notificationService.models.Notification;
import com.socialsphere.notificationService.repository.NotificationRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.PageRequest;
import org.springframework.http.ResponseEntity;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.*;

import java.security.Principal;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

@RestController
public class HttpController {
    @Autowired
    private NotificationRepository notificationRepository;
    @Autowired
    private SimpMessagingTemplate simpMessagingTemplate;

    @GetMapping("/api/v1/notifications/{id}")
    public ResponseEntity<NotificationDto> getNotification(@PathVariable Long id, Principal principal) throws JsonProcessingException {
        Notification notification = notificationRepository.getReferenceById(id);
        ObjectMapper mapper = new ObjectMapper();
        MessageDto messageDto = mapper.readValue(notification.getPayload(), MessageDto.class);
        NotificationDto resp = new NotificationDto(
                notification.getUserId(),
                messageDto.toString(),
                notification.getType(),
                notification.getRead(),
                notification.getCreatedAt()
        );
        return ResponseEntity.ok(resp);
    }

    @GetMapping("/api/v1/notifications/page")
    public ResponseEntity<List<NotificationDto>> getNotifications(@AuthenticationPrincipal UserDto user, @RequestParam(defaultValue = "0") int pageNo,
                                                                  @RequestParam(defaultValue = "10") int pageSize) {
        System.out.println("hit route");
        Optional<List<Notification>> notifications = notificationRepository.findByUserId((long) user.getUserId(), PageRequest.of(pageNo,pageSize));
        List<NotificationDto> notificationsList = notifications.orElseThrow().stream()
                .map(notification -> {
                    return new NotificationDto(notification.getUserId(), notification.getPayload(), notification.getType(), notification.getRead(), notification.getCreatedAt());
                })
                .collect(Collectors.toList());
        return ResponseEntity.ok(notificationsList);
    }

    @PostMapping("/api/v1/notifications/follow")
    public ResponseEntity createNotification(@RequestBody FollowDto followDto) {
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
