import json
import logging
import os
import boto3
import r2
import invoice_generator
from datetime import datetime, timezone

def update_invoice_pdf_url(invoice_id: str, pdf_url: str):
    """DynamoDBのpdf_urlフィールドを更新する"""
    dynamodb = boto3.resource('dynamodb', region_name=os.environ.get('AWS_REGION', 'ap-northeast-1'))
    table_name = os.environ.get('INVOICE_TABLE_NAME', 'incra-invoices')
    table = dynamodb.Table(table_name)
    table.update_item(
        Key={'invoice_id': invoice_id},
        UpdateExpression='SET pdf_url = :url, updated_at = :now',
        ExpressionAttributeValues={
            ':url': pdf_url,
            ':now': datetime.now(timezone.utc).isoformat()
        }
    )

def lambda_handler(event, context):
    logging.basicConfig(level=logging.INFO)
    logger = logging.getLogger()

    logger.info("Processing PDF generation request...")
    logger.info(f"Event: {json.dumps(event)}")

    for record in event.get('Records', []):
        try:
            invoice_data = json.loads(record['body'])
            logger.info(f"Processing invoice: {invoice_data.get('invoice_id')}")

            invoice_id = invoice_data.get('invoice_id', 'unknown')
            file_path = f"/tmp/{invoice_id}.pdf"

            # 品目データの変換
            items = invoice_data.get('items', [])
            invoice_details = []
            for item in items:
                invoice_details.append([
                    item.get('date', ''),
                    item.get('description', ''),
                    str(item.get('quantity', 0)),
                    str(item.get('unit_price', 0)),
                    str(item.get('amount', 0)),
                    item.get('memo', '')
                ])

            # デフォルト品目（品目がない場合）
            if not invoice_details:
                invoice_details = [
                    ['', '請求金額', '1', str(invoice_data.get('total_amount', 0)), str(invoice_data.get('total_amount', 0)), '']
                ]

            # PDF生成
            gen = invoice_generator.InvoiceGenerator()
            gen.generate_invoice(
                payer_name=invoice_data.get('billing_client_name', ''),
                sender_name=invoice_data.get('issuer_slack_real_name', ''),
                invoice_amount=invoice_data.get('total_amount', 0),
                deadline=invoice_data.get('due_date', ''),
                invoice_details=invoice_details,
                payment_method='銀行振込',
                payment_account=invoice_data.get('bank_details', ''),
                remarks=invoice_data.get('additional_info', ''),
                output_file=file_path
            )

            # Cloudflare R2 にアップロード
            upload_url = r2.upload_to_cloudflare(file_path)

            if upload_url:
                logger.info(f"PDF uploaded successfully: {upload_url}")
                # DynamoDBのpdf_urlを更新
                update_invoice_pdf_url(invoice_id, upload_url)
                logger.info(f"Updated DynamoDB pdf_url for invoice: {invoice_id}")
            else:
                logger.error(f"Failed to upload PDF for invoice: {invoice_id}")

        except Exception as e:
            logger.error(f"Error processing record: {e}")
            raise

    return {
        "statusCode": 200,
        "body": json.dumps({"message": "PDF generation completed"})
    }
