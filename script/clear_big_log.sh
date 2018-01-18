#!/bin/bash

cd /home/ec2-user/AdClick/Service/sengine
ls -lht |grep 4.0G|awk '{print $9}'|xargs rm
touch adx.log

cd /home/ec2-user/AdClick/Service/spostback
ls -lht |grep 4.0G|awk '{print $9}'|xargs rm
touch adx.log
