mongo <<EOF
use message_service
db.createUser({
  user: "mongo",
  pwd: "password",
  roles: [{ role: "readWrite", db: "message_service" }]
})
EOF