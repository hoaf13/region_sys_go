import requests
import time

url = 'http://127.0.0.1:8000/apis/v1/model/'

files = {'uploaded_file': open('/home/hoaf13/Downloads/sc.jpg','rb')}

start = time.time()
for i in range(10):
    print("iter: ", i)
    r = requests.post(url, files=files)
    print("response:", r)
end = time.time()

print("Total Time:", end - start, "second")
