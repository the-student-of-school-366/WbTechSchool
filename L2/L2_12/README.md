Чтобы проверить работу утилиты перейдите в диооекторию с файлом
```
cd L2\L2_12
```
Введите следующие команды:
```
go build -o grep.exe .
```

```
echo "hello`nworld`nhello again" | .\grep.exe hello
```

вы должны получить вывод:
```
hello
hello again
```