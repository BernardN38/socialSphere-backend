import pika
import time
import os
from io import BytesIO
from PIL import Image
import sys
from s3_helpers import get_image_from_s3, upload_image_to_s3
from mimetypes import guess_extension
import threading


def callback(ch, method, properties, body):
    # print(body, properties)
    try:
        image = Image.open(BytesIO(body))
        imageId = properties.headers.get('imageId')
        content_type = properties.headers.get('contentType')
        guess = guess_extension(content_type)
        print(guess)
        extension = guess.strip('.')
        compress_and_upload_image(image, imageId, extension)
    except Exception as e:
        print("exception", e )
    t = time.localtime()
    current_time = time.strftime("%H:%M:%S", t)
    print(current_time)
    print('image uploaded')
    ch.basic_ack(delivery_tag=method.delivery_tag)

def compress_and_upload_image(image, image_id, extension):
    if extension == 'jpg':
        extension = 'jpeg'
    img_io = BytesIO()
    print(image.width, image.height)
    if image.width > 1920 or image.height > 1080:
        new_width, new_height = 1920, 1080
        aspect_ratio = image.width / image.height
        if aspect_ratio > 1.777: # check if aspect ratio is wider than 16:9
            new_width = int(new_height * aspect_ratio)
        else:
            new_height = int(new_width / aspect_ratio)
        # Resize the image
        image = image.resize((new_width, new_height))
    if extension == 'png':
        # image = image.convert(mode='RGB')
        image = image.quantize(colors=256, method=2)
    image.save(img_io, 
                 extension, 
                 optimize = True, 
                 quality = 60)
    img_io.seek(0)
    time.sleep(2)
    # upload_image_to_s3(img_io, image_id)
    return

def worker(queue_name):
    connection = pika.BlockingConnection(pika.ConnectionParameters('rabbitmq'))
    channel = connection.channel()
    # channel.queue_declare(queue=queue_name)
    channel.basic_consume(queue_name, callback, auto_ack=False)
    channel.start_consuming()

def main():
    time.sleep(10)
    for i in range(10):
        t = threading.Thread(target=worker, args=("image-proccessing-worker",))
        t.start()
    t.join()
    
if __name__ == '__main__':
    try:
        main()
    except KeyboardInterrupt:
        print('Interrupted')
        try:
            sys.exit(0)
        except SystemExit:
            os._exit(0)