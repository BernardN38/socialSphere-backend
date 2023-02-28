package com.socialsphere.notificationService.configuration;

import com.socialsphere.notificationService.dto.UserDto;
import com.socialsphere.notificationService.util.JwtTokenProvider;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.GrantedAuthority;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.User;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.web.authentication.WebAuthenticationDetailsSource;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;


@Component
@AllArgsConstructor
@NoArgsConstructor
public class JwtAuthFilter extends OncePerRequestFilter {

    @Autowired
    private JwtTokenProvider jwtTokenProvider;

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain) throws ServletException, IOException {
        final String jwt = getCookie(request, "jwtToken");
        if (jwt == null) {
            System.out.println("jwt not found");
            filterChain.doFilter(request,response);
            return;
        }
        // print all request details
        System.out.println("Request URL: " + request.getRequestURL());
        System.out.println("Request Method: " + request.getMethod());
        System.out.println("Request Headers: " + request.getHeaderNames());
        System.out.println("Request Parameters: " + request.getParameterMap());
        System.out.println("Request Remote Address: " + request.getRemoteAddr());
        System.out.println("Request Session ID: " + request.getSession().getId());

        UserDto user = jwtTokenProvider.getUserFromToken(jwt);
        System.out.println(user.getUserId() + "---" + user.getUsername());
        List<GrantedAuthority> authorities
                = new ArrayList<>();
        authorities.add(new SimpleGrantedAuthority("ROLE_USER"));
        UsernamePasswordAuthenticationToken authToken = new UsernamePasswordAuthenticationToken(
                user,null, authorities
        );

        authToken.setDetails(new WebAuthenticationDetailsSource().buildDetails(request));
        SecurityContextHolder.getContext().setAuthentication(authToken);
        System.out.println( SecurityContextHolder.getContext().getAuthentication().isAuthenticated());
        System.out.println("auth success: "+authToken.getName());
        filterChain.doFilter(request,response);
    }

    private String getCookie(HttpServletRequest req, String name) {
        Cookie[] cookies = req.getCookies();
        if (cookies == null){
            return  null;
        }
        for (Cookie cookie : cookies) {
            if (cookie.getName().equals(name))
                return cookie.getValue();
        }
        return null;
    }
}