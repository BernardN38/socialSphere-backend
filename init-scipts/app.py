from sqlalchemy import Column, Integer,String, create_engine
from sqlalchemy.orm import declarative_base, Session
import bcrypt 

authentication_db = create_engine("postgresql://postgres:password@localhost:5438/authentication_service", echo=True, future=True)
identity_db = create_engine("postgresql://postgres:password@localhost:5438/identity_service", echo=True, future=True)
friends_db = create_engine("postgresql://postgres:password@localhost:5438/friend_service", echo=True, future=True)
image_db = create_engine("postgresql://postgres:password@localhost:5438/image_service", echo=True, future=True)
post_db = create_engine("postgresql://postgres:password@localhost:5438/post_service", echo=True, future=True)
Base = declarative_base()

class User(Base):
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


with Session(authentication_db) as session:
    
    password = "password"
    # converting password to array of bytes
    bytes = password.encode('utf-8')
  
    # generating the salt
    salt = bcrypt.gensalt(12)
  
    # Hashing the password
    hash = bcrypt.hashpw(bytes, salt)
    
    test_user = User(
        username="testUser1",
        email="test@gmail.com",
        password = hash.decode('utf-8'),
        first_name="testFirstName",
        last_name="testLastName"
    )

    session.add_all([test_user])

    session.commit()


	# Username  string `json:"username" validate:"required,min=2,max=100"`
	# Email     string `json:"email" validate:"required,email"`
	# Password  string `json:"password" validate:"required,min=8,max=128"`
	# FirstName string `json:"firstName" validate:"required"`
	# LastName  string `json:"lastName" validate:"required"`