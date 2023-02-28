package com.socialsphere.notificationService.util;

import com.socialsphere.notificationService.dto.UserDto;
import io.jsonwebtoken.Claims;
import io.jsonwebtoken.ExpiredJwtException;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.SignatureException;
import org.springframework.stereotype.Component;

import java.util.Base64;
import java.util.List;

@Component
public class JwtTokenProvider {

    private String secretKey= "qwertyuiopasdfghjklzxcvbnm123456";


    public boolean validateToken(String token) {
        try {
            Jwts.parser().setSigningKey(secretKey.getBytes()).parseClaimsJws(token);
            return true;
        } catch (SignatureException ex) {
            System.out.println(ex);
            System.out.println("signature invalid");
            // invalid signature
        } catch (ExpiredJwtException ex) {
            System.out.println("expired");
            // expired token
        } catch (Exception ex) {
            // other errors
            System.out.println(ex);
            System.out.println("other");
        }
        return false;
    }

    public UserDto getUserFromToken(String token) {
        Claims claims = Jwts.parser().setSigningKey(secretKey.getBytes()).parseClaimsJws(token).getBody();
        UserDto user = new UserDto(claims.getSubject(),Integer.parseInt(claims.getId()));
        return user;
    }

    public List<String> getRolesFromToken(String token) {
        Claims claims = Jwts.parser().setSigningKey(secretKey.getBytes()).parseClaimsJws(token).getBody();
        return claims.get("roles", List.class);
    }
}