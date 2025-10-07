package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type MoodleFetcher struct {
	client *http.Client

	user       string
	pass       string
	loginSite  string
	gradesSite string
}

func NewMoodleFetcher(LoginSite, gradesSite, user, pass string) *MoodleFetcher {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	return &MoodleFetcher{
		client: &http.Client{
			Jar: jar,
		},
		user:       user,
		pass:       pass,
		loginSite:  LoginSite,
		gradesSite: gradesSite,
	}
}

func (gp *MoodleFetcher) Login() error {
	loginPageResp, err := gp.client.Get(gp.loginSite)
	if err != nil {
		return fmt.Errorf("error fetching login page: %v", err)
	}
	defer loginPageResp.Body.Close()

	if loginPageResp.StatusCode != http.StatusOK {
		return fmt.Errorf("login page returned status: %s", loginPageResp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(loginPageResp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse login page HTML: %v", err)
	}

	form := doc.Find("form#login").First()
	if form.Length() == 0 {
		doc.Find("form").EachWithBreak(func(i int, s *goquery.Selection) bool {
			if s.Find("input[name='username']").Length() > 0 && s.Find("input[name='password']").Length() > 0 {
				form = s
				return false
			}
			return true
		})
	}

	if form.Length() == 0 {
		return errors.New("could not find login form on the page")
	}

	action, exists := form.Attr("action")
	if !exists {
		return errors.New("login form does not have an action attribute")
	}

	// Resolve relative action URL
	actionURL, err := url.Parse(action)
	if err != nil {
		return fmt.Errorf("invalid form action url: %v", err)
	}
	baseURL, _ := url.Parse(gp.loginSite)
	actionURL = baseURL.ResolveReference(actionURL)

	data := url.Values{}

	// collect form fields (including hidden fields like logintoken)
	form.Find("input").Each(func(i int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if !ok || strings.TrimSpace(name) == "" {
			return
		}
		typ, _ := s.Attr("type")
		if typ == "submit" || typ == "button" {
			return
		}
		val, _ := s.Attr("value")
		data.Set(name, val)
	})

	// set username/password
	data.Set("username", gp.user)
	data.Set("password", gp.pass)

	resp, err := gp.client.PostForm(actionURL.String(), data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body for quick checks
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	// Simple heuristics to determine success: Moodle redirects to /my/ or contains 'Dashboard' or 'Log out'
	if resp.Request != nil {
		finalURL := resp.Request.URL.String()

		if strings.Contains(finalURL, "/my/") || strings.Contains(strings.ToLower(bodyStr), "log out") || strings.Contains(strings.ToLower(bodyStr), "dashboard") {
			return nil
		}
	}

	// fallback: check body for error message
	if strings.Contains(strings.ToLower(bodyStr), "invalid") || strings.Contains(strings.ToLower(bodyStr), "incorrect") {
		return errors.New("login failed: invalid credentials or error present in response")
	}

	return err
}

func (gp *MoodleFetcher) GetGradesPage() ([]byte, error) {
	return gp.Fetch(gp.gradesSite)
}

func (gp *MoodleFetcher) Fetch(link string) ([]byte, error) {
	resp, err := gp.client.Get(link)
	if err != nil {
		return nil, fmt.Errorf("error fetching grades page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("grades page returned status: %s", resp.Status)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	return buf.Bytes(), err
}
