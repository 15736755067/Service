package request

import (
	"fmt"
	"testing"
)

var cacheStr = `YWZmSWQ9MTE3JmFmZk5hbWU9bW9iYWlyJmJvdD1mYWxzZSZicmFuZD1MZW5vdm8mYnJvd3Nlcj1DaHJvbWUmYnJvd3NlcnY9NTYuMC4yOTI0Ljg3JmNDb3VudHJ5PSZjSGFzaD1jOTgxZDIzMS0yODdmLTQwNTAtODBlNy05OGU4OTcxMDkyMWImY0lkPTEzMiZjTmFtZT1wb3BhZHMubmV0XzErLStJbmRpYSstK0luZGlhKy0rdWNfbmV3c19JTl9wcm9kdWN0aW9uJmNhcnJpZXI9VGhpcytwYXJhbWV0ZXIraXMrdW5hdmFpbGFibGUrZm9yK3NlbGVjdGVkK2RhdGErZmlsZS4rUGxlYXNlK3VwZ3JhZGUrdGhlK2RhdGErZmlsZS4mY2l0eT1UaGlzK3BhcmFtZXRlcitpcyt1bmF2YWlsYWJsZStmb3Irc2VsZWN0ZWQrZGF0YStmaWxlLitQbGVhc2UrdXBncmFkZSt0aGUrZGF0YStmaWxlLiZjbGlja1RzPTE0OTAyOTIyNTIyMzkmY29ublR5cGU9VGhpcytwYXJhbWV0ZXIraXMrdW5hdmFpbGFibGUrZm9yK3NlbGVjdGVkK2RhdGErZmlsZS4rUGxlYXNlK3VwZ3JhZGUrdGhlK2RhdGErZmlsZS4mY29zdD0wLjAwMTAwMCZjb3VudHJ5Q29kZT0mY291bnRyeU5hbWU9VGhpcytwYXJhbWV0ZXIraXMrdW5hdmFpbGFibGUrZm9yK3NlbGVjdGVkK2RhdGErZmlsZS4rUGxlYXNlK3VwZ3JhZGUrdGhlK2RhdGErZmlsZS4mY3BhVmFsdWU9MC4wMDAwMDAmZFR5cGU9TW9iaWxlJmV4dGVybmFsSWQ9NjkzODIxMTQ3MSZmSWQ9Mzc4JmZsb3dOYW1lPWRlZmF1bHROYW1lJmlkPTJlOGJiNTExODk3NWYxMGJkMDViYjYwMzQxNDI4MWI2MzQmaW1wVHM9MCZpcD0yNDA1JTNBMjA0JTNBZDMwOSUzQTkxYSUzQSUzQWIwMyUzQTE4YWQmaXNwPVRoaXMrcGFyYW1ldGVyK2lzK3VuYXZhaWxhYmxlK2ZvcitzZWxlY3RlZCtkYXRhK2ZpbGUuK1BsZWFzZSt1cGdyYWRlK3RoZStkYXRhK2ZpbGUuJmxJZD03OSZsTmFtZT1HbG9iYWwrLSt1Y19uZXdzX0lEXzIwMTcwMzE2MTQmbGFuZ3VhZ2U9ZW4tSU4mbW9kZWw9QTcwMjBhNDgmb0FmZklkPTExNyZvSWQ9MTAzJm9OYW1lPW1vYmFpcistK0luZGlhKy0rVUMrTmV3cyslMjhBbmRyb2lkJTI5JTI4Tm9uLWluY2VudCUyOSslMkZJTiZvT0lkPTEwMyZvcz1BbmRyb2lkKzYuMCZvc3Y9Ni4wJnBJZD0yMjUmcGF5b3V0PTAuNTAwMDAwJnBiVHM9MTQ5MDQ5NTY5MTI2OCZySWQ9MjA4JnJlZj0mcmVmRG9tYWluPSZyZWdpb249VGhpcytwYXJhbWV0ZXIraXMrdW5hdmFpbGFibGUrZm9yK3NlbGVjdGVkK2RhdGErZmlsZS4rUGxlYXNlK3VwZ3JhZGUrdGhlK2RhdGErZmlsZS4mdD1zMnNwb3N0YmFjayZ0cmtEb21haW49c2lybzdjLm5idHJrMC5jb20mdHJrUGF0aD0lMkZwb3N0YmFjayZ0c0NJZD00NDgxMjkwJnRzSWQ9MTMyJnRzTmFtZT1wb3BhZHMubmV0XzEmdHNWYXJzPSUzQiUzQSUzQSUzQTAlM0IlM0ElM0ElM0EwJTNCQURCTE9DSyUzQSU1QkFEQkxPQ0slNUQlM0FBREJMT0NLJTNBMSUzQkJST1dTRVJJRCUzQSU1QkJST1dTRVJJRCU1RCUzQUJST1dTRVJJRCUzQTElM0JCUk9XU0VSTkFNRSUzQSU1QkJST1dTRVJOQU1FJTVEJTNBQlJPV1NFUk5BTUUlM0ExJTNCQ0FNUEFJR05OQU1FJTNBJTVCQ0FNUEFJR05OQU1FJTVEJTNBQ0FNUEFJR05OQU1FJTNBMSUzQkNBVEVHT1JZSUQlM0ElNUJDQVRFR09SWUlEJTVEJTNBQ0FURUdPUllJRCUzQTElM0JDQVRFR09SWU5BTUUlM0ElNUJDQVRFR09SWU5BTUUlNUQlM0FDQVRFR09SWU5BTUUlM0ExJTNCQ09VTlRSWSUzQSU1QkNPVU5UUlklNUQlM0FDT1VOVFJZJTNBMSUzQkRFVklDRUlEJTNBJTVCREVWSUNFSUQlNUQlM0FERVZJQ0VJRCUzQTElM0JERVZJQ0VOQU1FJTNBJTVCREVWSUNFTkFNRSU1RCUzQURFVklDRU5BTUUlM0ExJTNCRk9STUZBQ1RPUklEJTNBJTVCRk9STUZBQ1RPUklEJTVEJTNBRk9STUZBQ1RPUklEJTNBMSZ0eElkPSZ1SWQ9MjQmdUlkVGV4dD1zaXJvN2MmdWE9TW96aWxsYSUyRjUuMCslMjhMaW51eCUzQitBbmRyb2lkKzYuMCUzQitMZW5vdm8rQTcwMjBhNDgrQnVpbGQlMkZNUkE1OEslMjkrQXBwbGVXZWJLaXQlMkY1MzcuMzYrJTI4S0hUTUwlMkMrbGlrZStHZWNrbyUyOStDaHJvbWUlMkY1Ni4wLjI5MjQuODcrTW9iaWxlK1NhZmFyaSUyRjUzNy4zNiZ2YXJzPTAlM0I4NTczJTNCR29vZ2xlK0Nocm9tZSslMkYrNTYlM0IrSW5kaWErLSt1Y19uZXdzX0lOJTNCJTNCJTNCQVAlM0I3MzE1JTNCTGVub3ZvKyUyRitWSUJFK0s1K05vdGUlM0IzMTAmdmlzaXRUcz0xNDkwMjkyMjQzMDQ1JndlYnNpdGVJZD0xODg4MTMx`

func TestAAA(t *testing.T) {
	req := CacheStr2Req(cacheStr)
	fmt.Println(req.RemoteIp())
}