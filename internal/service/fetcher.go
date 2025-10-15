package service

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
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/config"
	"golang.org/x/sync/singleflight"
)

var (
	ErrNotLogIn         = errors.New("❗️ not logged in")
	ErrWrongCredentials = errors.New("❗️ wrong credentials")
)

type MoodleFetcher struct {
	loginGroup singleflight.Group
	client     *http.Client

	user       string
	pass       string
	loginPage  string
	mainPage   string
	gradesPage string
}

func NewMoodleFetcher(cfg config.MoodleConfig) *MoodleFetcher {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	return &MoodleFetcher{
		loginGroup: singleflight.Group{},
		client: &http.Client{
			Jar: jar,
		},
		user:       cfg.MoodleUser,
		pass:       cfg.MoodlePass,
		loginPage:  cfg.MoodleLoginPage,
		gradesPage: cfg.MoodleGradePage,
		mainPage:   cfg.MoodleMainPage,
	}
}

func (gp *MoodleFetcher) IsLogined() error {
	resp, err := gp.client.Get(gp.mainPage)
	if err != nil {
		return fmt.Errorf("error checking login status: %v", err)
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse login page HTML: %v", err)
	}

	if doc.Find("a[href*='logout']").Length() > 0 {
		return nil
	}

	if doc.Url != nil && doc.Url.String() == gp.mainPage {
		return nil
	}

	return ErrNotLogIn
}
func (gp *MoodleFetcher) Login() error {
	_, err, _ := gp.loginGroup.Do("login", func() (interface{}, error) {

		loginPageResp, err := gp.client.Get(gp.loginPage)
		if err != nil {
			return nil, fmt.Errorf("error fetching login page: %v", err)
		}
		defer loginPageResp.Body.Close()

		if loginPageResp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("login page returned status: %s", loginPageResp.Status)
		}

		doc, err := goquery.NewDocumentFromReader(loginPageResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse login page HTML: %v", err)
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
			return nil, errors.New("could not find login form on the page")
		}

		action, exists := form.Attr("action")
		if !exists {
			return nil, errors.New("login form does not have an action attribute")
		}

		// Resolve relative action URL
		actionURL, err := url.Parse(action)
		if err != nil {
			return nil, fmt.Errorf("invalid form action url: %v", err)
		}

		baseURL, _ := url.Parse(gp.loginPage)
		actionURL = baseURL.ResolveReference(actionURL)

		data := url.Values{}

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
			return nil, err
		}
		defer resp.Body.Close()

		// Read response body for quick checks
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)

		if resp.Request != nil {
			finalURL := resp.Request.URL.String()

			if strings.Contains(finalURL, "/my/") || strings.Contains(strings.ToLower(bodyStr), "log out") || strings.Contains(strings.ToLower(bodyStr), "dashboard") {
				return nil, nil
			}
		}

		if strings.Contains(strings.ToLower(bodyStr), "invalid") || strings.Contains(strings.ToLower(bodyStr), "incorrect") {
			return nil, ErrWrongCredentials
		}

		return nil, err
	})

	return err
}

func (gp *MoodleFetcher) GetGradesPage() ([]byte, error) {
	return gp.Fetch(gp.gradesPage)
}

func (gp *MoodleFetcher) Fetch(link string) ([]byte, error) {
	resp, err := gp.client.Get(link)
	if err != nil {
		return nil, fmt.Errorf("error fetching grades page: %v", err)
	}
	defer resp.Body.Close()

	if resp.Request.URL.String() != link {

		err := gp.Login()
		if err != nil {
			return nil, fmt.Errorf("re-login failed: %v", err)
		}

		resp.Body.Close()
		resp, err = gp.client.Get(link)
		if err != nil {
			return nil, fmt.Errorf("error fetching grades page: %v", err)
		}
	}

	// slog.Debug("Fetched  page",
	// 	"status", resp.Status,
	// 	"url", resp.Request.URL.String(),
	// 	"link", link)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("grades page returned status: %s", resp.Status)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	return buf.Bytes(), err
}
