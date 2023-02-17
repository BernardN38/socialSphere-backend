package com.socialsphere.notificationService.RedisListener;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.socialsphere.notificationService.dto.NotificationDto;
import com.socialsphere.notificationService.models.Notification;
import com.socialsphere.notificationService.repository.NotificationRepository;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;
import org.springframework.data.redis.connection.Message;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.connection.MessageListener;
import org.springframework.messaging.simp.SimpMessageSendingOperations;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Service;

import java.sql.Timestamp;

@AllArgsConstructor
@NoArgsConstructor
@Service
public class RedisMessageSubscriber implements MessageListener {
    private NotificationRepository notificationRepository;

    private SimpMessagingTemplate simpMessagingTemplate;


    @Override
    public void onMessage(Message message, byte[] pattern) {
        String payload = new String(message.getBody());
        System.out.println(payload);
        ObjectMapper mapper = new ObjectMapper();
        NotificationDto notificationDto;
        try {
            notificationDto = mapper.readValue(payload, NotificationDto.class);
        } catch (JsonProcessingException e) {
            throw new RuntimeException(e);
        }
        Notification notification = new Notification(notificationDto.getUserId(), notificationDto.getPayload().toString(), notificationDto.getType());
        notificationRepository.save(notification);
        simpMessagingTemplate.convertAndSendToUser(notificationDto.getUserId().toString(), "/notifications", notification);
    }
}