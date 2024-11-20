from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import A4, portrait
from reportlab.lib import colors
from reportlab.platypus import Table, TableStyle
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.pdfbase import pdfmetrics
from datetime import datetime
import os

class Invoice:
    def __init__(self, payer_name, sender_name, invoice_amount, deadline, payment_method, payment_account, invoice_filename):
        self.payer_name = payer_name
        self.sender_name = sender_name
        self.invoice_amount = invoice_amount
        self.deadline = deadline
        self.payment_method = payment_method
        self.payment_account = payment_account
        self.invoice_filename = invoice_filename

        # フォントの登録
        pdfmetrics.registerFont(TTFont(os.getenv('FONT_NAME'), os.getenv('FONT_PATH')))

    def format_amount(self, amount):
        formatted_amount = f"{amount:,.0f}"
        return formatted_amount

    def format_date(self, yyyy_mm_dd: str) -> str:
        if len(yyyy_mm_dd) != 10 or yyyy_mm_dd[4] != '/' or yyyy_mm_dd[7] != '/':
            raise ValueError("Invalid date format. The correct format is yyyy/mm/dd.")
        try:
            year, month, day = map(int, yyyy_mm_dd.split('/'))
        except ValueError:
            raise ValueError("Invalid date components. Year, month, and day must be integers.")
        
        if not (1 <= month <= 12):
            raise ValueError("Month must be between 1 and 12.")
        
        try:
            datetime(year, month, day)  # 日付の存在チェック
        except ValueError:
            raise ValueError(f"Invalid day for the given month and year: {yyyy_mm_dd}")
        
        return f"{year}年{month}月{day}日"

    def gen_invoice(self):
        # PDFファイルを新規作成
        c = canvas.Canvas(self.invoice_filename, pagesize=portrait(A4))
        width, height = A4

        # タイトル: 請求書（中央に配置）
        c.setFont("IPAexGothic", 20)
        title = "請求書"
        title_width = c.stringWidth(title, "IPAexGothic", 20)  # 文字列の幅を取得
        c.drawString((width - title_width) / 2, height - 50, title)  # ページの中央に配置

        # 宛先: 誰々 御中
        c.setFont("IPAexGothic", 12)
        c.drawString(50, height - 100, self.payer_name + " 御中")

        # 文言: 下記の通り、ご請求申し上げます
        c.drawString(50, height - 130, "下記の通り、ご請求申し上げます")

        # ご請求金額
        c.setFont("IPAexGothic", 14)
        c.drawString(50, height - 180, "ご請求金額 ￥" + self.format_amount(self.invoice_amount))

        # ご請求金額の下に下線
        c.line(50, height - 185, 200, height - 185)

        # お支払い期限
        c.setFont("IPAexGothic", 12)
        c.drawString(50, height - 220, "お支払い期限: " + self.format_date(self.deadline))

        # 請求者情報（右寄せ）
        c.setFont("IPAexGothic", 18)
        c.drawRightString(width - 50, height - 180, self.sender_name)
        c.setFont("IPAexGothic", 12)

        # 振込先情報
        if self.payment_method:
            table_width = width - 100
            x_offset = 50
            y_offset = 350
            summary_data_2 = [
                [self.payment_method, self.payment_account]
            ]
            summary_table_2 = Table(summary_data_2, colWidths=[150, table_width - 150], rowHeights=45)

            summary_table_2.setStyle(TableStyle([
                ('BACKGROUND', (0, 0), (0, -1), colors.lightblue),
                ('TEXTCOLOR', (0, 0), (0, 0), colors.black),  # タイトルの文字色を黒に変更
                ('TEXTCOLOR', (1, 0), (1, -1), colors.black),  # 内容の文字色を黒に変更
                ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
                ('FONTNAME', (0, 0), (-1, -1), 'IPAexGothic'),
                ('BOTTOMPADDING', (0, 0), (-1, 0), 10),
                ('GRID', (0, 0), (-1, -1), 0.5, colors.black),
            ]))

            summary_table_2.wrapOn(c, table_width, height - y_offset - 250)
            summary_table_2.drawOn(c, x_offset, height - y_offset - 250)

        # PDFを保存
        c.showPage()
        c.save()

        print("PDFファイル '{}' が生成されました".format(self.invoice_filename))
