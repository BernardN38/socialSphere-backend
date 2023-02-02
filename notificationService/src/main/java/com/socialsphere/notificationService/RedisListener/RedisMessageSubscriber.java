package com.socialsphere.notificationService.RedisListener;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.socialsphere.notificationService.dto.NotificationDto;
import com.socialsphere.notificationService.models.Notification;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;
import org.springframework.data.redis.connection.Message;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.connection.MessageListener;
import org.springframework.messaging.simp.SimpMessageSendingOperations;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Component;
import org.springframework.stereotype.Controller;
import org.springframework.stereotype.Service;

@AllArgsConstructor
@NoArgsConstructor
@Service
public class RedisMessageSubscriber implements MessageListener {


    private SimpMessagingTemplate simpMessagingTemplate;
    @Override
    public void onMessage(Message message, byte[] pattern) {
        String payload = new String(message.getBody());
        ObjectMapper mapper = new ObjectMapper();
        NotificationDto notification;
        try {
            notification = mapper.readValue(payload, NotificationDto.class);
        } catch (JsonProcessingException e) {
            throw new RuntimeException(e);
        }
        simpMessagingTemplate.convertAndSendToUser(notification.getUserId().toString(), "/notifications", notification);
    }
}