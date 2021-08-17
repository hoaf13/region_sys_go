import pika
import time
import random
import redis 
import os 

red = redis.StrictRedis(host='localhost')

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host='localhost',port=5672))
channel = connection.channel()

channel.queue_declare(queue='task_queue', durable=True)
print(' [*] Waiting for messages. To exit press CTRL+C')


def callback(ch, method, properties, body):
    print("worker: ", os.getpid())
    print("body: ", body)
    body = body.decode("utf-8") 
    data = body.split('|')
    uploaded_file = data[1]
    uploaded_time = data[0]
    print(" [x] Received {}".format(uploaded_file))
    delay = random.randint(1,10)
    time.sleep(0.5) 
    ans = random.choice(["north", "central", "south"])
    red.set(uploaded_time, str(ans))    
    print(f"upload time: {uploaded_time} - label: {ans}")
    print(" [x] Done\n\n")
    ch.basic_ack(delivery_tag=method.delivery_tag)


channel.basic_qos(prefetch_count=1) # no more than 1 message in each consumerng
channel.basic_consume(queue='task_queue', on_message_callback=callback)
channel.start_consuming()

