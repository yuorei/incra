from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import A4, portrait
from reportlab.lib import colors
from reportlab.platypus import Table, TableStyle
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.pdfbase import pdfmetrics
from datetime import datetime
from typing import List
import os

class InvoiceGenerator:
    def __init__(self):
        self.font = os.getenv('FONT_NAME')
        self.font_path =os.getenv('FONT_PATH')
        print(self.font, self.font_path,"フォントを読み込みました")
        
        pdfmetrics.registerFont(TTFont(self.font, self.font_path))

    @staticmethod
    def format_amount(amount: int) -> str:
        return f"{amount:,.0f}"

    @staticmethod
    def format_date(yyyy_mm_dd: str) -> str:
        try:
            year, month, day = map(int, yyyy_mm_dd.split('/'))
            if not 1 <= month <= 12 or not 1 <= day <= 31:
                raise ValueError("Invalid date format. Use yyyy/mm/dd.")
            if month in [4, 6, 9, 11] and day > 30:
                raise ValueError("Invalid date format. Use yyyy/mm/dd.")
            if month == 2 and day > 29:
                raise ValueError("Invalid date format. Use yyyy/mm/dd.")
            return f"{year}年{month}月{day}日"
        except ValueError:
            raise ValueError("Invalid date format. Use yyyy/mm/dd.")

    def generate_invoice(self, payer_name: str, sender_name: str, invoice_amount: int, deadline: str, 
                         invoice_details: List[List[str]], payment_method: str, payment_account: str, 
                         remarks: str, output_file: str = "invoice.pdf"):
        c = canvas.Canvas(output_file, pagesize=portrait(A4))
        width, height = A4

        # Title
        c.setFont(self.font, 20)
        title = "請求書"
        title_width = c.stringWidth(title, self.font, 20)
        c.drawString((width - title_width) / 2, height - 50, title)

        # Payer
        c.setFont(self.font, 12)
        c.drawString(50, height - 100, f"{payer_name} 御中")
        c.drawString(50, height - 130, "下記の通り、ご請求申し上げます")

        # Invoice Amount
        c.setFont(self.font, 14)
        c.drawString(50, height - 180, f"ご請求金額 ￥{self.format_amount(invoice_amount)}")
        c.line(50, height - 185, 200, height - 185)

        # Payment Deadline
        c.setFont(self.font, 12)
        c.drawString(50, height - 220, f"お支払い期限: {self.format_date(deadline)}")

        # Sender Info
        c.setFont(self.font, 18)
        c.drawRightString(width - 50, height - 180, sender_name)

        # Table Data
        header = ["取引日付", "内容", "数量", "単価", "金額", "概要"]
        data = [header] + invoice_details
        x_offset, y_offset = 50, 350
        table_width = width - 100
        details_table = Table(data, colWidths=[70, 100, 50, 70, 70, 100], rowHeights=25)
        details_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.lightblue),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.black),
            ('TEXTCOLOR', (0, 1), (-1, -1), colors.black),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, -1), self.font),
            ('GRID', (0, 0), (-1, -1), 0.5, colors.black),
        ]))
        if invoice_details:
            details_table.wrapOn(c, table_width, y_offset)
            details_table.drawOn(c, x_offset, height - y_offset)

        # Total Amount
        total_amount = sum(int(row[4].replace(",", "")) for row in invoice_details)
        total_table = Table(
            [["合計金額", f"￥{self.format_amount(total_amount)}"]],
            colWidths=[150, table_width - 150],
            rowHeights=30
        )
        total_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (0, -1), colors.lightblue),
            ('TEXTCOLOR', (0, 0), (0, 0), colors.black),
            ('TEXTCOLOR', (1, 0), (1, -1), colors.black),
            ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
            ('FONTNAME', (0, 0), (-1, -1), self.font),
            ('GRID', (0, 0), (-1, -1), 0.5, colors.black),
        ]))
        if invoice_details:
            total_table.wrapOn(c, table_width, height - y_offset - 180)
            total_table.drawOn(c, x_offset, height - y_offset - 180)

        # Payment Information
        payment_table = Table(
            [[payment_method, payment_account]],
            colWidths=[150, table_width - 150],
            rowHeights=30
        )
        payment_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (0, -1), colors.lightblue),
            ('TEXTCOLOR', (0, 0), (0, 0), colors.black),
            ('TEXTCOLOR', (1, 0), (1, -1), colors.black),
            ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
            ('FONTNAME', (0, 0), (-1, -1), self.font),
            ('GRID', (0, 0), (-1, -1), 0.5, colors.black),
        ]))
        payment_table.wrapOn(c, table_width, height - y_offset - 250)
        payment_table.drawOn(c, x_offset, height - y_offset - 250)

        # Remarks
        if remarks:
            remarks_table = Table(
                [["備考", remarks]],
                colWidths=[150, table_width - 150],
                rowHeights=45
            )
            remarks_table.setStyle(TableStyle([
                ('BACKGROUND', (0, 0), (0, -1), colors.lightblue),
                ('TEXTCOLOR', (0, 0), (0, 0), colors.black),
                ('TEXTCOLOR', (1, 0), (1, -1), colors.black),
                ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
                ('FONTNAME', (0, 0), (-1, -1), self.font),
                ('GRID', (0, 0), (-1, -1), 0.5, colors.black),
            ]))
            remarks_table.wrapOn(c, table_width, height - y_offset - 320)
            remarks_table.drawOn(c, x_offset, height - y_offset - 320)

        # Save PDF
        c.showPage()
        c.save()
        
        print(datetime.now().strftime("%Y/%m/%d %H:%M:%S"), "PDFファイル", output_file, "が作成されました")

# Usage Example
if __name__ == "__main__":
    invoice_generator = InvoiceGenerator()
    invoice_generator.generate_invoice(
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
        output_file="invoice.pdf"
    )
    
