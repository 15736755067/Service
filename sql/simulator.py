#!/usr/bin/python
# encoding=utf8

"""
模拟traffic source请求

需求
1. 支持设置比率（impression to campaign, campaign to click, click to postback)
2. 支持设置同时多少个线程（进程）并行请求。
3. 同一个线程的请求，是序列化的。
4. 支持统计信息：类似于ab的输出
5. 支持一些错误的检查
6. 只支持对一个campaign的测试，暂时不支持同时测试多个campaign
7. clickid 从重定向的链接里面获取
8. 记录下来请求的数量，方便查看。
9. 从重定向链接中，分析是lander page还是offer。如果是lander page才需要模拟click
"""

import argparse
import logging
import requests
import base64
import re
from urllib import unquote

import sys

from timer import Timer

# http://chenxing.mymac.com:52500/campaign.1.hash

# 参数解析
parser = argparse.ArgumentParser()
parser.add_argument("--loglevel", default="DEBUG", help="ERROR,WARN,INFO,DEBUG")
parser.add_argument("--campaign", default="463d-be39-e1f51d80591", help="the campaign you want to test")
parser.add_argument("--host", default="aabbcc.hemin.im", help="service host")
parser.add_argument("--port", default=52500, type=int, help="service port")
parser.add_argument("--count", default=1, type=int, help="how many requests to send to the campaign")
parser.add_argument("--phost", default="aabbcc.hemin.im", help="postback host")
parser.add_argument("--pport", default=55500, type=int, help="postback port")
parser.add_argument("--impression", default=False, type=bool, help="send impression request or not")
parser.add_argument("--multi", default="thread", type=str, help="multi thread or multi process")

args = parser.parse_args()
logging.basicConfig(level=args.loglevel)

if args.multi == "thread":
    from multiprocessing.dummy import Pool as ThreadPool
else:
    from multiprocessing import Pool as ThreadPool

campaign_url = "http://%s:%s/%s" % (args.host, args.port, args.campaign)
logging.debug("campaign_url: %s", campaign_url)

impression_url = "http://%s:%s/impression/%s" % (args.host, args.port, args.campaign)
logging.debug("impression_url: %s", impression_url)

click_url = "http://%s:%s/click" % (args.host, args.port)
postback_url_template = "http://%s:%s/postback?cid=%s&payout=%s&txid=%s" % (
    args.phost, args.pport, "%s", "%s", "%s"
)


def get_postback_url(cid, payout, txid):
    return postback_url_template % (cid, payout, txid)


def parse_tstep(decoded):
    """
    parse_tstep 解析tstep里面的内容
    :param decoded: reqId=ba8bf5e4debe131f69bfbd9e0ef91e7a&step=lpofferreq
    :return: {'reqId': 'ba8bf5e4debe131f69bfbd9e0ef91e7a', 'step': 'lpofferreq'}
    """
    m = {}
    for kv in decoded.split("&"):
        k, v = kv.split("=")
        m[k] = v
    return m


def lander_or_offer(url):
    if url.find("lander") != -1:
        return "lander"
    return "offer"


def parse_redirect_url(req):
    """
    parse_redirect_url try to parse redirect url from request
    :param req: http request
    :return: redirect url, None if error
    """
    if req is None:
        return

    if req.status_code == 302:
        # 302直接跳转
        return req.headers["location"]

    if req.status_code == 200:
        # meta refresh 或者 double meta refresh
        g = re.search('url=(.*?)"', req.text)
        if g is None:
            logging.error("bad req.text:%s", req.text)
            return

        url = g.group(1)
        if url.startswith('/dmr?dest='):
            # <meta http-equiv="refresh" content="0;url=/dmr?dest=http%3A%2F%2Fdownload.androidapp.baidu.com%2Fpublic%2Fuploads%2Fstore_0%2Fstatic%2Fmobisummer4_gl_bd_apk-com.duapps.cleaner-1220_mobisummer4_gl_bd_apk.apk%3Fcid%3Df5db45f0c0a14ade3d43e26918a6d2c6">
            # double meta refresh
            dest_url_quoted = url[len('/dmr?dest='):]
            dest_url = unquote(dest_url_quoted)
            return dest_url
        else:
            # <meta http-equiv="refresh" content="0;url=http://download.androidapp.baidu.com/public/uploads/store_0/static/mobisummer4_gl_bd_apk-com.duapps.cleaner-1220_mobisummer4_gl_bd_apk.apk?cid=3be7c017f2a83b64929d56b7b2edc21f">
            return g.group(1)

    logging.error("parse_redirect_url failed status_code:%s req.text:%s", req.status_code, req.text)
    return


def parse_clickid(redirect):
    g = re.search("[^a-zA-Z0-9]cid=([a-zA-Z0-9]+)", redirect)
    return g.group(1)


REQ_NO = 0
REQ_SUCCESS = 1
REQ_FAILED = 2

statusToS = {
    REQ_NO: "none",
    REQ_SUCCESS: "success",
    REQ_FAILED: "failed",
}


class OneReqStat(object):
    def __init__(self):
        self.Status = REQ_NO
        self.Time = 0.0

    def __str__(self):
        return "%s:%s" % (statusToS[self.Status], self.Time)


class ReqSerialStat(object):
    def __init__(self):
        self.Imp = OneReqStat()
        self.Camp = OneReqStat()
        self.Click = OneReqStat()
        self.Postback = OneReqStat()

    def __repr__(self):
        return "<Imp:%s Camp:%s Click:%s Postback:%s>" % (
            self.Imp,
            self.Camp,
            self.Click,
            self.Postback,
        )


# OneReqStat = namedtuple('OneReqStat', [
#     'ImpressionCnt', 'ImpressionTime',
#     'CampaignCnt',  'ClickCnt', 'PostbackCnt'], verbose=True)


def send_request(url, stat, prefix, cookies=None):
    with Timer() as t:
        try:
            impression = requests.get(url, cookies=cookies, allow_redirects=False)
        except requests.ConnectionError as ce:
            print >>sys.stderr, ce
            return

    stat.Time = t.interval

    logging.debug("%s: status_code: %s", prefix, impression.status_code)
    logging.debug("%s: headers: %s", prefix, impression.headers)
    logging.debug("%s: encoding: %s", prefix, impression.encoding)
    logging.debug("%s: cookies: %s", prefix, impression.cookies)

    return impression


# run_once 模拟一次从impression 到 campaign 到 click 到 postback请求
def run_once(i):
    stat = ReqSerialStat()

    jar = requests.cookies.RequestsCookieJar()

    logging.debug("%s%s%s", '*' * 40, "%dth" % i, '*' * 40)
    # impression模拟
    if args.impression:
        impression = send_request(impression_url, stat=stat.Imp, prefix="impression.%d" % i, cookies=jar)
        if impression and impression.status_code == 200:
            logging.debug("impression: text: %s", impression.text)
            stat.Imp.Status = REQ_SUCCESS
            jar = impression.cookies
        else:
            stat.Imp.Status = REQ_FAILED

    logging.debug('-' * 80)

    # campaign模拟
    req = send_request(campaign_url, stat=stat.Camp, prefix="campaign.%d" % i, cookies=jar)
    if req and req.status_code == 200:
        logging.debug("campaign.%d text: [%s]", i, req.text)

    # 三种情况：1. 302直接跳转   2. meta refresh跳转   3. double meta refresh
    redirect_url = parse_redirect_url(req)
    if redirect_url is None:
        # 重定向链接错误，服务器逻辑有问题
        logging.error("campaign.%d request redirect url error", i)
        stat.Camp.Status = REQ_FAILED
        return stat

    stat.Camp.Status = REQ_SUCCESS
    logging.debug("redirect url: %s", redirect_url)

    clickid = parse_clickid(redirect_url)
    logging.debug("clickid %s", clickid)

    # click 模拟
    # 如果直接走到offer，就没有click这一步了
    logging.debug('-' * 80)
    if lander_or_offer(redirect_url) == "lander":
        click = send_request(click_url, stat=stat.Click, prefix="click.%d" % i, cookies=req.cookies)

        if click and click.status_code == 200:
            logging.debug("click text: %s", click.text)
        lander_redirect_url = parse_redirect_url(click)
        logging.debug("click.%d lander_redirect_url:%s", i, lander_redirect_url)
        if lander_redirect_url is None:
            stat.Click.Status = REQ_FAILED
        else:
            stat.Click.Status = REQ_SUCCESS
        logging.debug('-' * 80)

    # postback 模拟1
    postback_url = get_postback_url(clickid, 0.01, "this-is-transaction-id")
    postback = send_request(postback_url, stat=stat.Postback, prefix="postback.%d" % i)

    logging.debug('-' * 80)
    if postback and postback.status_code == 200:
        stat.Postback.Status = REQ_SUCCESS
    else:
        stat.Postback.Status = REQ_FAILED

    return stat


pool = ThreadPool()
results = pool.map(run_once, xrange(args.count))


def output(name, stat_getter):
    success = 0
    failed = 0
    total_time = 0.0
    min_time = 60.0
    max_time = 0.0
    avg_time = 0.0

    for req_serial in results:
        s = stat_getter(req_serial)

        if s.Status == REQ_SUCCESS:
            success += 1
        elif s.Status == REQ_FAILED:
            failed += 1

        if s.Status != REQ_NO:
            total_time += s.Time
            if s.Time > max_time:
                max_time = s.Time
            if s.Time < min_time:
                min_time = s.Time

    total = success + failed
    if total > 0:
        avg_time = total_time / total

    print name, "success:", success, "failed:", failed, "avg time:", avg_time, "max_time:", max_time, "min_time:", min_time


output("impression", lambda serial: serial.Imp)
output("campaign", lambda serial: serial.Camp)
output("click", lambda serial: serial.Click)
output("postback", lambda serial: serial.Postback)

# do clean up
pool.close()
pool.join()
