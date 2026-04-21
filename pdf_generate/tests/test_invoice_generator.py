import os
import sys
import unittest
import tempfile
from unittest.mock import patch

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))

from invoice_generator import _sanitize, InvoiceGenerator


class TestSanitize(unittest.TestCase):
    def test_ascii(self):
        self.assertEqual(_sanitize("hello"), "hello")

    def test_japanese(self):
        self.assertEqual(_sanitize("山田太郎"), "山田太郎")

    def test_emoji_removed(self):
        self.assertEqual(_sanitize("Tomo🎵田中"), "Tomo田中")

    def test_multiple_emoji_removed(self):
        self.assertEqual(_sanitize("🌟abc🎉"), "abc")

    def test_empty(self):
        self.assertEqual(_sanitize(""), "")

    def test_mixed_symbols(self):
        self.assertEqual(_sanitize("株式会社ABC / 田中"), "株式会社ABC / 田中")


class TestFormatAmount(unittest.TestCase):
    def test_thousands(self):
        self.assertEqual(InvoiceGenerator.format_amount(1000), "1,000")

    def test_zero(self):
        self.assertEqual(InvoiceGenerator.format_amount(0), "0")

    def test_large(self):
        self.assertEqual(InvoiceGenerator.format_amount(1000000), "1,000,000")


class TestFormatDate(unittest.TestCase):
    def test_slash_format(self):
        self.assertEqual(InvoiceGenerator.format_date("2024/12/31"), "2024年12月31日")

    def test_hyphen_format(self):
        self.assertEqual(InvoiceGenerator.format_date("2026-04-20"), "2026年4月20日")

    def test_invalid_month(self):
        with self.assertRaises(ValueError):
            InvoiceGenerator.format_date("2024/13/01")

    def test_invalid_day(self):
        with self.assertRaises(ValueError):
            InvoiceGenerator.format_date("2024/04/31")

    def test_invalid_feb(self):
        with self.assertRaises(ValueError):
            InvoiceGenerator.format_date("2024/02/30")


class TestGenerateInvoice(unittest.TestCase):
    def setUp(self):
        font_path = os.path.join(os.path.dirname(__file__), '..', 'src', 'ipam.ttf')
        self._env_patcher = patch.dict(os.environ, {'FONT_NAME': 'IPAexMincho', 'FONT_PATH': font_path})
        self._env_patcher.start()
        self.gen = InvoiceGenerator()

    def tearDown(self):
        self._env_patcher.stop()

    def test_generates_pdf_file(self):
        with tempfile.NamedTemporaryFile(suffix='.pdf', delete=False) as f:
            output = f.name
        try:
            self.gen.generate_invoice(
                payer_name="株式会社テスト",
                sender_name="山田太郎",
                invoice_amount=10000,
                deadline="2024/12/31",
                invoice_details=[["2024/11/01", "商品A", "1", "10,000", "10,000", ""]],
                payment_method="銀行振込",
                payment_account="〇〇銀行 普通 1234567",
                remarks="",
                output_file=output,
            )
            self.assertTrue(os.path.exists(output))
            self.assertGreater(os.path.getsize(output), 0)
        finally:
            if os.path.exists(output):
                os.remove(output)

    def test_generates_pdf_with_emoji_in_name(self):
        with tempfile.NamedTemporaryFile(suffix='.pdf', delete=False) as f:
            output = f.name
        try:
            self.gen.generate_invoice(
                payer_name="Tomo🎵田中",
                sender_name="送信者🌟",
                invoice_amount=1000,
                deadline="2026-04-20",
                invoice_details=[["2026-04-20", "キャンセル料", "1", "1,000", "1,000", ""]],
                payment_method="銀行振込",
                payment_account="ことら送金",
                remarks="",
                output_file=output,
            )
            self.assertTrue(os.path.exists(output))
        finally:
            if os.path.exists(output):
                os.remove(output)

    def test_generates_pdf_without_remarks(self):
        with tempfile.NamedTemporaryFile(suffix='.pdf', delete=False) as f:
            output = f.name
        try:
            self.gen.generate_invoice(
                payer_name="テスト株式会社",
                sender_name="送信者",
                invoice_amount=5000,
                deadline="2024/06/30",
                invoice_details=[["2024/06/01", "サービス料", "1", "5,000", "5,000", "備考あり"]],
                payment_method="銀行振込",
                payment_account="△△銀行 普通 9999999",
                remarks="",
                output_file=output,
            )
            self.assertTrue(os.path.exists(output))
        finally:
            if os.path.exists(output):
                os.remove(output)


if __name__ == '__main__':
    unittest.main()
