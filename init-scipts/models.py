from sqlalchemy import Column, Integer,String
from sqlalchemy.orm import declarative_base, Session
Base = declarative_base()


class AuthUser(Base):
     __tablename__ = "users"

     id = Column(Integer, primary_key=True)
     username = Column(String)
     email = Column(String)
     password = Column(String)
     first_name = Column(String)
     last_name = Column(String)

     def __repr__(self):
         return f"User(id={self.id!r}, name={self.name!r}, fullname={self.fullname!r})"

	# Username  string `json:"username" validate:"required,min=2,max=100"`
	# Email     string `json:"email" validate:"required,email"`
	# Password  string `json:"password" validate:"required,min=8,max=128"`
	# FirstName string `json:"firstName" validate:"required"`
	# LastName  string `json:"lastName" validate:"required"`
