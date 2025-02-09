import os
from typing import Optional
import boto3
from botocore.config import Config

s3 = boto3.client(
    "s3",
    endpoint_url=os.getenv("R2_ENDPOINT_URL"),
    aws_access_key_id=os.getenv("AWS_ACCESS_KEY_ID"),
    aws_secret_access_key=os.getenv("AWS_SECRET_ACCESS_KEY"),
    config=Config(signature_version="s3v4"),
    region_name=os.getenv("REGION_NAME"),
)

BUCKET_NAME =os.getenv("BUCKET_NAME") 
CLOUDFLARE_PUBLIC_URL = ""

def upload_to_cloudflare(file_path: str, object_name: Optional[str] = None) -> str:
    """
    ファイルをCloudflare R2にアップロードし、公開URLを返し、ローカルファイルを削除します。

    :param file_path: アップロードするファイルへのパス
    :param object_name: S3オブジェクト名。指定されない場合、file_pathのbasenameが使用されます
    :return: アップロードされたファイルの公開URL
    """
    # S3 object_nameが指定されていない場合、file_pathのbasenameを使用
    if object_name is None:
        object_name = os.path.basename(file_path)

    try:
        # ファイルをアップロード
        s3.upload_file(file_path, BUCKET_NAME, object_name)
        
        # アップロードされたファイルの公開URLを生成
        url = f"{CLOUDFLARE_PUBLIC_URL}{object_name}"
        
        # ローカルファイルを削除
        os.remove(file_path)
        
        return url
    except Exception as e:
        print(f"エラーが発生しました: {e}")
        return ""
