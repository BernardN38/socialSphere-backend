import bcrypt
from sqlalchemy.orm import Session
from sqlalchemy import create_engine
from models import AuthUser


authentication_db = create_engine(
    "postgresql://postgres:password@localhost:5438/authentication_service", echo=True, future=True)
identity_db = create_engine(
    "postgresql://postgres:password@localhost:5438/identity_service", echo=True, future=True)
friends_db = create_engine(
    "postgresql://postgres:password@localhost:5438/friend_service", echo=True, future=True)
image_db = create_engine(
    "postgresql://postgres:password@localhost:5438/image_service", echo=True, future=True)
post_db = create_engine(
    "postgresql://postgres:password@localhost:5438/post_service", echo=True, future=True)


password = "password"
# converting password to array of bytes
bytes = password.encode('utf-8')

# generating the salt
salt = bcrypt.gensalt(12)

# Hashing the password
hash = bcrypt.hashpw(bytes, salt)


users = [
    {"username": "erisd1",
        "email": "ericd@gmail.com",
        "password": hash.decode('utf-8'),
        "first_name": "eric",
        "last_name": "diaz"},
    {
        "username": "ednap2",
        "email": "ednap@gmail.com",
        "password": hash.decode('utf-8'),
        "first_name": "edna",
        "last_name": "pina"
    },
    {
        "username": "erisboxx",
        "email": "erisboxx@gmail.com",
        "password": hash.decode('utf-8'),
        "first_name": "bernardn",
        "last_name": "narvaez"
    },
    {"username": "rbobby",
        "email": "rickybobyreyes@gmail.com",
        "password": hash.decode('utf-8'),
        "first_name": "ricky",
        "last_name": "reyes"
     },
    {"username": "cflores3",
     "email": "cflores93@gmail.com",
     "password": hash.decode('utf-8'),
     "first_name": "chris",
     "last_name": "flores"
     }
]


def create_auth_users():
    with Session(authentication_db) as session:

        test_user = AuthUser(
        username=users[0]["username"],
        email=users[0]["email"],
        password=hash.decode('utf-8'),
        first_name=users[0]["first_name"],
        last_name=users[0]["last_name"]
    )
        test_user2 = AuthUser(
        username=users[1]["username"],
        email=users[1]["email"],
        password=hash.decode('utf-8'),
        first_name=users[1]["first_name"],
        last_name=users[1]["last_name"]
    )
        test_user3 = AuthUser(
        username=users[2]["username"],
        email=users[2]["email"],
        password=hash.decode('utf-8'),
        first_name=users[2]["first_name"],
        last_name=users[2]["last_name"]
    )
        test_user4 = AuthUser(
        username=users[3]["username"],
        email=users[3]["email"],
        password=hash.decode('utf-8'),
        first_name=users[3]["first_name"],
        last_name=users[3]["last_name"]
    )
        test_user5 = AuthUser(
        username=users[4]["username"],
        email=users[4]["email"],
        password=hash.decode('utf-8'),
        first_name=users[4]["first_name"],
        last_name=users[4]["last_name"]
    )

        session.add_all(
        [  test_user, test_user2, test_user3, test_user4, test_user5])

        session.commit()

def create_identity_users():
    with Session(identity_db) as session:

        test_user = AuthUser(
        username=users[0]["username"],
        email=users[0]["email"],
        password=hash.decode('utf-8'),
        first_name=users[0]["first_name"],
        last_name=users[0]["last_name"]
    )
        test_user2 = AuthUser(
        username=users[1]["username"],
        email=users[1]["email"],
        password=hash.decode('utf-8'),
        first_name=users[1]["first_name"],
        last_name=users[1]["last_name"]
    )
        test_user3 = AuthUser(
        username=users[2]["username"],
        email=users[2]["email"],
        password=hash.decode('utf-8'),
        first_name=users[2]["first_name"],
        last_name=users[2]["last_name"]
    )
        test_user4 = AuthUser(
        username=users[3]["username"],
        email=users[3]["email"],
        password=hash.decode('utf-8'),
        first_name=users[3]["first_name"],
        last_name=users[3]["last_name"]
    )
        test_user5 = AuthUser(
        username=users[4]["username"],
        email=users[4]["email"],
        password=hash.decode('utf-8'),
        first_name=users[4]["first_name"],
        last_name=users[4]["last_name"]
    )

        session.add_all(
        [  test_user, test_user2, test_user3, test_user4, test_user5])

        session.commit()
    return
