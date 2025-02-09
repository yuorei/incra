import json
import logging
import r2
import invoice_generator
import os

def lambda_handler(event, context):
    logging.basicConfig(level=logging.INFO)
    logger = logging.getLogger()
    
    logger.info("Processing request...")
    print(event)
    
    # 一時ファイルのパス
    file_path = "/tmp/invoice.pdf"

    invoice_generator_instance = invoice_generator.InvoiceGenerator()
    invoice_generator_instance.generate_invoice(
        payer_name="株式会社〇〇",
        sender_name="株式会社△△",
        invoice_amount=100000,
        deadline="2024/12/31",
        invoice_details=[
            ["2024/11/01", "商品A", "10", "1,000", "10,000", "サンプル内容"],
            ["2024/11/02", "商品B", "5", "2,000", "10,000", "サンプル内容"],
        ],
        payment_method="銀行振込",
        payment_account="〇〇銀行 △△支店 普通 1234567",
        remarks="お支払いに関するご質問はご連絡ください。",
        output_file=file_path
    )
        
    # Cloudflare R2 にアップロード
    upload_url = r2.upload_to_cloudflare(file_path)
    
    if upload_url:
        response_message = {"message": "File uploaded successfully!", "url": upload_url}
    else:
        response_message = {"message": "File upload failed."}
    
    return {
        "statusCode": 200,
        "body": json.dumps(response_message)
    }
