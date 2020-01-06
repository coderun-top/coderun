#!/usr/bin/env python
# -*- coding: UTF-8 -*-


import sys
import getopt
import urllib
import traceback
import datetime
import hashlib
import hmac
import subprocess
import random

HOST_HEADER = 'host'
DATE_HEADER = 'x-163-date'
SIG_VERSION_HEADER = 'x-163-signatureversion'
NONCE_HEADER = 'x-163-signaturenonce'
CONTENT_TYPE = 'content-type'
DRYRUN = 'x-163-dryrun'

SIGNED_HEADERS = [HOST_HEADER, DATE_HEADER, SIG_VERSION_HEADER,
                  NONCE_HEADER, CONTENT_TYPE, DRYRUN
                 ]

SIGNATURE_METHOD = 'HMAC-SHA256'

def get_signed_header_keys(headers):
    signed_header_keys = []
    for key in SIGNED_HEADERS:
        if key in headers:
            signed_header_keys.append(key)
    signed_header_keys = sorted(signed_header_keys)
    return signed_header_keys

def build_canonical_request(method, uri, query, headers, body):
    if '//' in uri:
        uri = uri[uri.index('//') + 2:]
    if '/' in uri:
        uri = uri[uri.index('/'):]
    else:
        uri = ''
    canonical_request = method.upper() + '\n'
    canonical_request += uri + '\n'
    canonical_request += build_params(query) + '\n'

    signed_header_keys = get_signed_header_keys(headers)
    signed_headers = {}
    for k in signed_header_keys:
        signed_headers[k] = headers[k]

    canonical_request += build_canonical_header(signed_headers) + '\n'
    canonical_request += build_signed_headers(signed_header_keys) + '\n'
    canonical_request += build_req_payload(body)

    return canonical_request

def build_params(query):
    keys = []
    res_map = {}
    for key in query:
        tmp_o = {}
        tmp_o[key] = query[key]
        tmp_s = urllib.urlencode(tmp_o)
        res_map[key] = tmp_s
        keys.append(key)
    keys = sorted(keys)
    # generate
    res = ''
    is_first = True
    for key in keys:
        if is_first:
            is_first = False
        else:
            res += '&'
        res += res_map[key]
    return res

def build_req_payload(body):
    s = ''
    if body is not None:
        s = body
    return hashlib.sha256(s).hexdigest()

def build_signed_headers(keys):
    return ';'.join(keys)

def build_canonical_header(headers):
    s = ''
    keys = []
    for k in headers:
        keys.append(k)
    keys = sorted(keys)
    for k in keys:
        v = headers[k]
        v = format_header_value(v)
        s += k + ':' + v + '\n'
    return s

def format_header_value(s):
    res = ''
    last_is_space = False
    for c in s:
        if c == ' ':
            if last_is_space:
                pass # ignore
            else:
                last_is_space = True
                res += c
        else:
            last_is_space = False
            res += c
    return res

def get_host_from_uri(uri):
    # remove http or https protocol
    if uri.startswith('http://') or uri.startswith('https://'):
        if uri.startswith('http://'):
            uri = uri[len('http://'):]
        else:
            uri = uri[len('https://'):]
    if '/' in uri:
        uri = uri[0:uri.index('/')]
    arr = uri.split(':')
    if len(arr) > 2:
        raise Exception('Malformed url: more than one `:`')
    elif len(arr) == 2 or len(arr) == 1:
        host = arr[0]
        return host
    else:
        raise Exception('Malformed url: ' + uri)

def split_url(url):
    arr = url.split('?')
    if len(arr) == 2:
        uri = arr[0]
        query = urllib.unquote(arr[1])
        arr2 = query.split('&')
        query_obj = {}
        for kv in arr2:
            kv_arr = kv.split('=')
            k = kv_arr[0]
            v = kv_arr[1]
            if len(kv_arr) == 2:
                if k in query_obj:
                    raise Exception('Malformed url: key exists more than once:' + k)
                query_obj[k] = v
            else:
                raise Exception('Malformed url: invalid key and value pair: ' + kv)
        return uri, query_obj
    elif len(arr) == 1:
        return arr[0], {}
    else:
        raise Exception('Malformed url: more than one `?`')

def build_string_to_sign(region, service, date, hex_cr):
    date_time_str = date.strftime('%Y-%m-%dT%H:%M:%SZ')

    string_to_sign = SIGNATURE_METHOD + '\n'
    string_to_sign += date_time_str + '\n'
    string_to_sign += get_credential_scope(date, region, service) + '\n'
    string_to_sign += hex_cr

    return string_to_sign

def get_credential_scope(date, region, service):
    date_time = date.strftime('%Y%m%d')
    return date_time + '/' + region + '/' + service + '/163_request'

def build_authorization(key, scope, headers, sig):
    keys = get_signed_header_keys(headers)
    return SIGNATURE_METHOD + ' Credential=' + key + '/' + scope + ', SignedHeaders=' + build_signed_headers(keys) + ', Signature=' + sig

def print_help():
    print ('Usage: ourl [options...] <url>')
    print ('Options:')
    print (' -d, --data DATA        \t\tHTTP POST data')
    print (' -H, --header LINE      \t\tPass custom header LINE to server')
    print (' -h, --help             \t\tThis help text')
    print (' -X, --request COMMAND  \t\tSpecify request command to use')
    print (' -v, --verbose          \t\tMake the operation more talkative')
    print (' -V                     \t\tShow ourl verbose info')
    print ('     --region           \t\tService region')
    print ('     --service          \t\tService identifier')
    print ('     --access-key       \t\tAccess key')
    print ('     --access-secret    \t\tAccess secret')

def gen_curl(verbose, method, url, headers, body):
    args = ['curl']
    if verbose:
        args.append('-v')
    args.append('-X')
    args.append(method)
    for h in headers:
        args.append('-H')
        args.append(h + ':' + headers[h])
    if body is not None:
        args.append('--data')
        args.append(body)
    args.append(url)
    return args

def call(args):
    p = subprocess.Popen(args)
    pout,perr = p.communicate('')
    if pout is not None:
        out = pout.decode('UTF-8')
        print out
    if perr is not None:
        err = perr.decode('UTF-8')
        print >> sys.stderr, err
    return p.returncode

def main():
    opts, args = getopt.getopt(sys.argv[1:], 'hvH:X:d:V', ['help', 'verbose', 'header=', 'request=', 'data=', 'region=', 'service=', 'access-key=', 'access-secret='])
    if len(args) != 1:
        print_help()
        return 1

    verbose = False
    curl_verbose = False
    headers = {}
    method = 'GET'
    body = None
    url = args[0]
    region = None
    service = None
    key = None
    secret = None

    for k, v in opts:
        if k in ('-h', '--help'):
            print_help()
            return 0
        elif k in ('-v', '--verbose'):
            curl_verbose = True
        elif k in ('-V'):
            verbose = True
        elif k in ('-H', '--header'):
            if ':' in v and v.index(':') < len(v) - 1:
                header_key = v[0:v.index(':')].strip().lower()
                header_value = v[v.index(':') + 1:].strip()
                headers[header_key] = header_value
            else:
                print >> sys.stderr, 'Malformed header: ' + v
                return 1
        elif k in ('-X', '--request'):
            method = v.upper()
        elif k in ('-d', '--data'):
            body = v
        elif k in ('--region'):
            region = v
        elif k in ('--service'):
            service = v
        elif k in ('--access-key'):
            key = v
        elif k in ('--access-secret'):
            secret = v
        else:
            print ('Unknown parameter ' + k)
            print_help()
            return 1

    printer = VerbosePrinter(verbose)

    if region is None:
        print >> sys.stderr, 'Region is not set: --region=...'
        return 1
    if service is None:
        print >> sys.stderr, 'Service is not set: --service=...'
        return 1
    if key is None:
        print >> sys.stderr, 'AccessKey is not set: --access-key=...'
        return 1
    if secret is None:
        print >> sys.stderr, 'AccessSecret is not set: --access-secret=...'
        return 1

    printer.show_var('method', method)
    printer.show_var('headers', headers)
    printer.show_var('url', url)
    printer.show_var('body', body)
    printer.show_var('region', region)
    printer.show_var('service', service)
    printer.show_var('access-key', key)
    printer.show_var('access-secret', secret)

    try:
        uri, query = split_url(url)
        printer.show_var('uri', uri)
        printer.show_var('query', query)

        # add headers if not present
        if SIG_VERSION_HEADER not in headers:
            headers[SIG_VERSION_HEADER] = '2.0'
        if NONCE_HEADER not in headers:
            rand = str(int(random.random() * 10000))
            headers[NONCE_HEADER] = rand
        if HOST_HEADER not in headers:
            host = get_host_from_uri(uri)
            headers[HOST_HEADER] = host
        if DATE_HEADER not in headers:
            now = datetime.datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%SZ')
            headers[DATE_HEADER] = now

        printer.show_var('headers_after_processing', headers)

        cr = build_canonical_request(method, uri, query, headers, body)
        printer.show_var('canonical_request', cr)

        hex_cr = hashlib.sha256(cr).hexdigest()
        printer.show_var('hased_canonical_request', hex_cr)

        # get time from header
        str_header_date = headers[DATE_HEADER]
        header_date = datetime.datetime.strptime(str_header_date, '%Y-%m-%dT%H:%M:%SZ')
        # str to sign
        str2sign = build_string_to_sign(region, service, header_date, hex_cr)
        printer.show_var('str2sign', str2sign)

        # hmac
        _d = header_date.strftime('%Y%m%d')
        printer.show_msg('HMAC(HMAC(HMAC(HMAC(163 + ' + secret + ', ' + _d + '), ' + region + '), ' + service + '), 163_request)')
        signing_key = HMAC(HMAC(HMAC(HMAC("163" + secret, _d), region), service), "163_request")
        printer.show_var('signing_key', signing_key)

        # sig
        sig = HexHMAC(signing_key, str2sign)

        # authorization header
        credential_scope = get_credential_scope(header_date, region, service)
        authorization = build_authorization(key, credential_scope, headers, sig)
        printer.show_var('authorization', authorization)

        # add header
        headers['Authorization'] = authorization

        # request
        curl_cmd = gen_curl(curl_verbose, method, url, headers, body)
        printer.show_var('curl', curl_cmd)
        return call(curl_cmd)
    except Exception as err:
        print >> sys.stderr, err.message
        if verbose:
            traceback.print_exc()
        return 1

class VerbosePrinter:
    def __init__(self, verbose):
        self.verbose = verbose
    def show_var(self, key, value):
        if self.verbose:
            print key + ' = ' + str(value)
    def show_msg(self, msg):
        if self.verbose:
            print msg

def HMAC(key, msg):
    return hmac.new(key, msg, hashlib.sha256).digest()

def HexHMAC(key, msg):
    return hmac.new(key, msg, hashlib.sha256).hexdigest()

# start
if __name__ == "__main__":
    sys.exit(main())
