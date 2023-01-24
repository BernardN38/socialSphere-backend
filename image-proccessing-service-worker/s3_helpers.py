from minio import Minio
import boto3

minioClient = Minio('minio:9000',
                  access_key='minio',
                  secret_key='minio123',
                  secure=False)


def get_image_from_s3(image_id):
    object1 = minioClient.get_object('image-service-socialsphere1', image_id)
    return object1

def upload_image_to_s3(image_obj, image_id):
    # bucket.upload_fileobj(image_obj, image_id)  
    minioClient.put_object('image-service-socialsphere1', image_id, image_obj, -1, part_size=5*1024*1024)
    return 