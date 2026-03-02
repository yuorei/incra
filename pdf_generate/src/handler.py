import json
import logging
import os
import invoice_generator
from slack_sdk import WebClient
from slack_sdk.errors import SlackApiError


def lambda_handler(event, context):
    logging.basicConfig(level=logging.INFO)
    logger = logging.getLogger()

    logger.info("Processing PDF generation request...")
    logger.info(f"Event: {json.dumps(event)}")

    slack_client = WebClient(token=os.environ.get('SLACK_TOKEN'))

    for record in event.get('Records', []):
        try:
            message = json.loads(record['body'])
            invoice_data = {k: v for k, v in message.items() if k != 'billing_client_slack_user_id'}
            billing_client_slack_user_id = message.get('billing_client_slack_user_id', '')
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

            # Slack DMでPDFファイルを送信
            if billing_client_slack_user_id:
                try:
                    slack_client.files_upload_v2(
                        channel=billing_client_slack_user_id,
                        file=file_path,
                        title=f"請求書 {invoice_id}",
                        initial_comment=f"請求書 {invoice_id} のPDFです。",
                    )
                    logger.info(f"PDF sent via Slack DM to {billing_client_slack_user_id} for invoice: {invoice_id}")
                except SlackApiError as e:
                    logger.error(f"Failed to send PDF via Slack DM: {e}")
                    raise
            else:
                logger.warning(f"No billing client Slack user ID for invoice: {invoice_id}, skipping Slack DM")

            # ローカルファイルのクリーンアップ
            if os.path.exists(file_path):
                os.remove(file_path)

        except Exception as e:
            logger.error(f"Error processing record: {e}")
            raise

    return {
        "statusCode": 200,
        "body": json.dumps({"message": "PDF generation completed"})
    }
