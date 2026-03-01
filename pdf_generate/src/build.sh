zip -r python_lambda.zip handler.py r2.py python/* ipam.ttf invoice_generator.py
rm ../terraform/lambda/python_lambda.zip
mv python_lambda.zip ../terraform/lambda
rm python_lambda.zip
