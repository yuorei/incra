zip -r python_lambda.zip handler.py python/* ipam.ttf invoice_generator.py
mkdir -p ../../infra/environments/prod/lambda
rm -f ../../infra/environments/prod/lambda/python_lambda.zip
mv python_lambda.zip ../../infra/environments/prod/lambda/
