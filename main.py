import os
import re
import requests
from datetime import datetime, timedelta
from slack_sdk import WebClient
from slack_sdk.errors import SlackApiError
from slack_sdk.signature import SignatureVerifier
from flask import Flask, request, jsonify
from invoice_generator import Invoice  # invoice_generator.py

class SlackEventProcessor:
    def __init__(self, slack_token, signing_secret):
        self.client = WebClient(token=slack_token)
        self.signature_verifier = SignatureVerifier(signing_secret)
        self.slack_token = slack_token

    def verify_signature(self, data, headers):
        """署名検証を行う"""
        return self.signature_verifier.is_valid_request(data, headers)

    def get_user_info(self, user_id):
        """ユーザー情報を取得する"""
        url = f"https://slack.com/api/users.info?user={user_id}"
        headers = {'Authorization': f'Bearer {self.slack_token}'}
        
        response = requests.get(url, headers=headers)
        
        if response.status_code == 200:
            user_json = response.json()
            if user_json.get('ok'):
                return user_json['user']['profile']['display_name']
        print(f"Error: Unable to retrieve user info (Status Code: {response.status_code})")
        return None

    def generate_invoice(self, user_id, channel, sender_name, invoice_amount, deadline, payment_method, payment_account):
        """請求書を生成し、ファイルをアップロードする"""
        now = datetime.now()
        invoice_filename = f"invoice_{user_id}_{user_id}_{now.strftime('%Y-%m-%d-%H-%M-%S')}.pdf"
        
        invoice = Invoice(
            payer_name=user_id,  # Display name is set to user_id here. Replace as needed.
            sender_name=sender_name,
            invoice_amount=invoice_amount,
            deadline=deadline,
            payment_method=payment_method,
            payment_account=payment_account,
            invoice_filename=invoice_filename
        )
        invoice.gen_invoice()

        try:
            response = self.client.files_upload_v2(
                file=invoice_filename,
                title=invoice_filename,
                channel=channel,
                initial_comment=user_id + "さんが、"+sender_name+"さんに請求書を送信しました。"
            )
            os.remove(invoice_filename)  # アップロード後にファイルを削除
        except SlackApiError as e:
            print(f"Error posting message: {e.response['error']}")
    

class SlackApp:
    def __init__(self, slack_token, signing_secret, bot_user_id):
        self.app = Flask(__name__)
        self.processor = SlackEventProcessor(slack_token, signing_secret)
        self.bot_user_id = bot_user_id

    def route(self):
        """Slackのイベントを処理するルート"""
        @self.app.route("/slack/events", methods=["POST"])
        def slack_events():
            data = request.json
            headers = request.headers

            # 署名検証
            if not self.processor.verify_signature(request.get_data(), headers):
                return "Request signature verification failed", 400

            # URL検証
            if data.get("type") == "url_verification":
                return jsonify({"challenge": data.get("challenge")})

            event = data.get('event', {})

            if event.get("type") == "app_mention":
                user_id = event.get("user")
                text = event.get("text")
                channel = event.get("channel")

                # メンションしたユーザーの情報を取得
                display_name = self.processor.get_user_info(user_id)
                if not display_name:
                    return jsonify({"status": "error", "message": "Failed to get user display name"})

                pattern = r'.*?<@' + re.escape(BOT_USER_ID) + r'>'
                text = re.sub(pattern, '', text)  # BotのIDを取り除く
                mentioned_user_ids = re.findall(r'<@(\w+)>', text)  # メンションされたユーザーのIDを抽出
                
                # 1人のユーザーがメンションされた場合、そのユーザーの情報も取得
                sender_name = None
                if mentioned_user_ids:
                    sender_user_id = mentioned_user_ids[0]
                    sender_name = self.processor.get_user_info(sender_user_id)
                    if not sender_name:
                        return jsonify({"status": "error", "message": "Failed to get mentioned user display name"})

                # メッセージをスペースで分割し、空白を取り除く
                text_parts = [item for item in text.split(' ') if item]
                invoice_amount = int(text_parts[1]) if len(text_parts) > 1 else 0
                deadline = text_parts[2] if len(text_parts) > 2 else (datetime.now() + timedelta(hours=9, days=7)).strftime("%Y/%m/%d")
                payment_method = text_parts[3] if len(text_parts) > 3 else ""
                payment_account = text_parts[4] if len(text_parts) > 4 else ""

                # 請求書生成
                self.processor.generate_invoice(user_id, channel, sender_name or text_parts[0], invoice_amount, deadline, payment_method, payment_account)

            return jsonify({"status": "ok"})

    def run(self):
        """Flaskアプリを実行"""
        self.route()
        self.app.run(port=3000)


if __name__ == "__main__":
    SLACK_TOKEN=os.getenv('SLACK_BOT_TOKEN')
    SLACK_SIGNING_SECRET=os.getenv('SLACK_SIGNING_SECRET')
    BOT_USER_ID =os.getenv('BOT_USER_ID')
    slack_app = SlackApp(SLACK_TOKEN, SLACK_SIGNING_SECRET, BOT_USER_ID)
    slack_app.run()
