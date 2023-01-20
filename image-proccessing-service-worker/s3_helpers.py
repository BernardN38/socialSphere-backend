
import boto3

s3 = boto3.resource('s3', region_name='us-east-2')

bucket = s3.Bucket('image-service-socialsphere1')


def get_image_from_s3(image_id):
    object1 = bucket.Object(image_id)
    return object1

def upload_image_to_s3(image_obj, image_id):
    bucket.upload_fileobj(image_obj, image_id)  
    return 