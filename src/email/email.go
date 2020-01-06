// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package email

import (
	tlspkg "crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

)

// Send ...
func Send(addr, identity, username, password string,
	timeout int, tls, insecure bool, from string,
	to []string, subject, message string) error {

	client, err := newClient(addr, identity, username,
		password, timeout, tls, insecure)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, t := range to {
		if err = client.Rcpt(t); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	template := "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\n%s\r\n"
	data := fmt.Sprintf(template, from,
		strings.Join(to, ","), subject, message)

	_, err = w.Write([]byte(data))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// Ping tests the connection and authentication with email server
// If tls is true, a secure connection is established, or Ping
// trys to upgrate the insecure connection to a secure one if
// email server supports it.
// Ping doesn't verify the server's certificate and hostname when
// needed if the parameter insecure is ture
func Ping(addr, identity, username, password string,
	timeout int, tls, insecure bool) error {
	client, err := newClient(addr, identity, username, password,
		timeout, tls, insecure)
	if err != nil {
		return err
	}
	defer client.Close()
	return nil
}

// caller needs to close the client
func newClient(addr, identity, username, password string,
	timeout int, tls, insecure bool) (*smtp.Client, error) {
	fmt.Printf("establishing TCP connection with %s ...", addr)
	conn, err := net.DialTimeout("tcp", addr,
		time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	if tls {
		fmt.Printf("establishing SSL/TLS connection with %s ...", addr)
		tlsConn := tlspkg.Client(conn, &tlspkg.Config{
			ServerName:         host,
			InsecureSkipVerify: insecure,
		})
		if err = tlsConn.Handshake(); err != nil {
			return nil, err
		}

		conn = tlsConn
	}

	fmt.Printf("creating SMTP client for %s ...", host)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return nil, err
	}

	//try to swith to SSL/TLS
	if !tls {
		if ok, _ := client.Extension("STARTTLS"); ok {
			fmt.Printf("switching the connection with %s to SSL/TLS ...", addr)
			if err = client.StartTLS(&tlspkg.Config{
				ServerName:         host,
				InsecureSkipVerify: insecure,
			}); err != nil {
				return nil, err
			}
		} else {
			fmt.Printf("the email server %s does not support STARTTLS", addr)
		}
	}

	if ok, _ := client.Extension("AUTH"); ok {
		fmt.Printf("authenticating the client...")
		// only support plain auth
		if err = client.Auth(smtp.PlainAuth(identity,
			username, password, host)); err != nil {
			return nil, err
		}
	} else {
		fmt.Printf("the email server %s does not support AUTH, skip",
			addr)
	}

	fmt.Printf("create smtp client successfully")

	return client, nil
}



//func (e *EmailAPI) Ping() {
//	var host, username, password, identity string
//	var port int
//	var ssl, insecure bool
//	body := e.Ctx.Input.CopyBody(1 << 32)
//	if body == nil || len(body) == 0 {
//		cfg, err := config.Email()
//		if err != nil {
//			log.Errorf("failed to get email configurations: %v", err)
//			e.CustomAbort(http.StatusInternalServerError,
//				http.StatusText(http.StatusInternalServerError))
//		}
//		host = cfg.Host
//		port = cfg.Port
//		username = cfg.Username
//		password = cfg.Password
//		identity = cfg.Identity
//		ssl = cfg.SSL
//		insecure = cfg.Insecure
//	} else {
//		settings := &struct {
//			Host     string  `json:"email_host"`
//			Port     *int    `json:"email_port"`
//			Username string  `json:"email_username"`
//			Password *string `json:"email_password"`
//			SSL      bool    `json:"email_ssl"`
//			Identity string  `json:"email_identity"`
//			Insecure bool    `json:"email_insecure"`
//		}{}
//		e.DecodeJSONReq(&settings)
//
//		if len(settings.Host) == 0 || settings.Port == nil {
//			e.CustomAbort(http.StatusBadRequest, "empty host or port")
//		}
//
//		if settings.Password == nil {
//			cfg, err := config.Email()
//			if err != nil {
//				log.Errorf("failed to get email configurations: %v", err)
//				e.CustomAbort(http.StatusInternalServerError,
//					http.StatusText(http.StatusInternalServerError))
//			}
//
//			settings.Password = &cfg.Password
//		}
//
//		host = settings.Host
//		port = *settings.Port
//		username = settings.Username
//		password = *settings.Password
//		identity = settings.Identity
//		ssl = settings.SSL
//		insecure = settings.Insecure
//	}
//
//	addr := net.JoinHostPort(host, strconv.Itoa(port))
//	if err := email.Ping(addr, identity, username,
//		password, pingEmailTimeout, ssl, insecure); err != nil {
//		log.Errorf("failed to ping email server: %v", err)
//		// do not return any detail information of the error, or may cause SSRF security issue #3755
//		e.RenderError(http.StatusBadRequest, "failed to ping email server")
//		return
//	}
//}
