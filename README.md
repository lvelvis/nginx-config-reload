# nginx-config-reload  
1.nginx-configmap配置热加载  
2.sidecar方式共享process、volume 

#测试  
``` 
### 1.创建nginx configmap  
kubectl create -f k8s-nginx-test/nginx-config-test.yaml

### 2.部署nginx
kubectl create -f k8s-nginx-test/test-nginx.yaml

### 3.修改configmap:nginx-config-v1 
kubectl edit configmap nginx-config-v1  

return 200 'ok 2020';   --替换--> return 200 '2020 hello';  

### 4.验证configmap是否生效  
while true;do curl http://nginx-svc-ip/healthz  && date -R ;done
```

