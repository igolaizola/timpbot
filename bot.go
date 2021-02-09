package bot

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
)

// Book creates a reservation
func Book(email, pass, center, activity, date, hour string) error {
	if err := check(center, activity, date); err != nil {
		return err
	}

	// Set cookiejar options
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		return fmt.Errorf("couldn't create cookiejar: %w", err)
	}
	client := &http.Client{
		Jar: jar,
	}

	if err := login(client, email, pass); err != nil {
		return err
	}

	if err := create(client, center, activity, date, hour); err != nil {
		return err
	}
	return nil
}

func check(center, activity, date string) error {
	// Check date
	u := fmt.Sprintf("https://connect.timp.pro/%s/activities/%s/admissions?date=%s", center, activity, date)
	checkRes, err := http.Get(u)
	if err != nil {
		return err
	}
	defer checkRes.Body.Close()
	if checkRes.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", checkRes.StatusCode, checkRes.Status)
	}
	checkDoc, err := goquery.NewDocumentFromReader(checkRes.Body)
	if err != nil {
		return err
	}
	var activeDate string
	checkDoc.Find("a.date-active").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		split := strings.Split(link, "=")
		if len(split) < 2 {
			return
		}
		activeDate = split[1]
	})
	if activeDate != date {
		return errors.New("date not found")
	}
	return nil
}

func login(client *http.Client, email, pass string) error {
	// Login
	loginReq, err := http.NewRequest("GET", "https://connect.timp.pro/login", nil)
	if err != nil {
		return fmt.Errorf("couldn't create request: %w", err)
	}
	loginRes, err := client.Do(loginReq)
	if err != nil {
		return err
	}
	defer loginRes.Body.Close()
	if loginRes.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", loginRes.StatusCode, loginRes.Status)
	}
	loginDoc, err := goquery.NewDocumentFromReader(loginRes.Body)
	if err != nil {
		return err
	}

	var crsfToken string
	loginDoc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		content, _ := s.Attr("content")
		if name == "csrf-token" {
			crsfToken = content
		}
	})
	if crsfToken == "" {
		return errors.New("crsf-token not found")
	}

	// Create session
	u := "https://connect.timp.pro/sessions"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("authenticity_token", crsfToken)
	_ = writer.WriteField("email", email)
	_ = writer.WriteField("password", pass)
	_ = writer.WriteField("permanent_session", "0")
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("couldn't close writer: %w", err)
	}

	sessionReq, err := http.NewRequest("POST", u, payload)
	if err != nil {
		return fmt.Errorf("couldn't create request: %w", err)
	}
	sessionReq.Header.Set("Content-Type", writer.FormDataContentType())
	sessionRes, err := client.Do(sessionReq)
	if err != nil {
		return fmt.Errorf("session request failed: %w", err)
	}

	defer sessionRes.Body.Close()
	if sessionRes.StatusCode != 200 {
		return fmt.Errorf("session status code error: %d %s", sessionRes.StatusCode, sessionRes.Status)
	}

	return nil
}

func create(client *http.Client, center, activity, date, hour string) error {
	// search id
	u := fmt.Sprintf("https://connect.timp.pro/%s/activities/%s/admissions?date=%s", center, activity, date)
	searchReq, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return fmt.Errorf("couldn't create request: %w", err)
	}
	searchRes, err := client.Do(searchReq)
	if err != nil {
		return err
	}
	defer searchRes.Body.Close()
	if searchRes.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", searchRes.StatusCode, searchRes.Status)
	}
	searchDoc, err := goquery.NewDocumentFromReader(searchRes.Body)
	if err != nil {
		return err
	}

	var crsfToken string
	searchDoc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		content, _ := s.Attr("content")
		if name == "csrf-token" {
			crsfToken = content
		}
	})
	if crsfToken == "" {
		return errors.New("crsf-token not found")
	}

	var id string
	searchDoc.Find("a.text-decoration-none.text-reset").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		split := strings.Split(strings.Trim(href, "/"), "/")
		if len(split) != 2 {
			return
		}
		currID := split[1]
		s.Find("div.p-3.text-center").Each(func(i int, s2 *goquery.Selection) {
			if s3 := s2.Find("div").First(); s3 != nil {
				if s3.Text() == hour {
					id = currID
				}
			}
		})
	})

	// book
	u = fmt.Sprintf("https://connect.timp.pro/admissions/%s/tickets?branch_building_id=%s", id, center)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("X-CSRF-Token", crsfToken)
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("couldn't close writer: %w", err)
	}

	bookReq, err := http.NewRequest("POST", u, payload)
	if err != nil {
		return fmt.Errorf("couldn't create request: %w", err)
	}
	bookReq.Header.Set("Content-Type", writer.FormDataContentType())
	bookReq.Header.Set("X-CSRF-Token", crsfToken)

	bookRes, err := client.Do(bookReq)
	if err != nil {
		return fmt.Errorf("book request failed: %w", err)
	}

	defer bookRes.Body.Close()
	if bookRes.StatusCode != 200 {
		return fmt.Errorf("book status code error: %d %s", bookRes.StatusCode, bookRes.Status)
	}
	return nil
}
