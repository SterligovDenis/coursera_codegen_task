package main
import "encoding/json"
import "strconv"
import "strings"
import "net/http"

	
	func (this *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		
			if r.URL.Path == "/user/profile" {
				this.handlerProfile(w, r)
				return
			}
		
			if r.URL.Path == "/user/create" {
				this.handlerCreate(w, r)
				return
			}
		
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown method\"}"))
	}

	
		func (this *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
			
	
			if r.Method == http.MethodPost {
				r.ParseForm()
			}
			
			
			
				
				
			
				
				
					v := ProfileParams{}
					
						
							if r.Method == http.MethodPost {
								v.Login = r.FormValue("login")
							} else {
								v.Login = r.URL.Query().Get("login")
							}
						

						
							
								if v.Login == "" {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"login must me not empty\"}"))
									return
								}
							
							
						

						
						

						

						
						
					
				
			
			ctx := r.Context()
			res, err := this.Profile(ctx, v)
			if err != nil {
				if apiErr, ok := err.(ApiError); ok {
					w.WriteHeader(apiErr.HTTPStatus)
					w.Write([]byte("{\"error\": \"" + apiErr.Error() + "\"}"))
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("{\"error\": \"" + err.Error() + "\"}"))
				}
				return
			}
			resMap := make(map[string]interface{})
			resMap["error"] = interface{}("")
			resMap["response"] = interface{}(res)
			jsonStr, err := json.Marshal(resMap)
			w.Write(jsonStr)	
		}
	
		func (this *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
			
				if r.Method != "POST" {
					w.WriteHeader(http.StatusNotAcceptable)
					w.Write([]byte("{\"error\": \"bad method\"}"))
					return
				}
			
	
			if r.Method == http.MethodPost {
				r.ParseForm()
			}
			
			
				if r.Header.Get("X-Auth") != "100500" {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte("{\"error\": \"unauthorized\"}"))
					return
				}
			
			
				
				
			
				
				
					v := CreateParams{}
					
						
							if r.Method == http.MethodPost {
								v.Login = r.FormValue("login")
							} else {
								v.Login = r.URL.Query().Get("login")
							}
						

						
							
								if v.Login == "" {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"login must me not empty\"}"))
									return
								}
							
							
						

						
						

						
							
								if len(v.Login) < 10 {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"login len must be >= 10\"}"))
									return
								}
							
						

						
						
					
						
							if r.Method == http.MethodPost {
								v.Name = r.FormValue("full_name")
							} else {
								v.Name = r.URL.Query().Get("full_name")
							}
						

						
							
							
						

						
						

						

						
						
					
						
							if r.Method == http.MethodPost {
								v.Status = r.FormValue("status")
							} else {
								v.Status = r.URL.Query().Get("status")
							}
						

						
							
							
								if v.Status == "" {
									v.Status = "user"
								}
							
						

						
						
							inEnum := false
							var errArr []string
							
							
							
							 	errArr = append(errArr, "user")
								if v.Status == "user" {
									inEnum = true
								}
							
							 	errArr = append(errArr, "moderator")
								if v.Status == "moderator" {
									inEnum = true
								}
							
							 	errArr = append(errArr, "admin")
								if v.Status == "admin" {
									inEnum = true
								}
								
							if !inEnum {
								w.WriteHeader(http.StatusBadRequest)
								w.Write([]byte("{\"error\": \"status must be one of ["+strings.Join(errArr, ", ")+"]\"}"))
								return
							}
						

						

						
						
					
						
							var val int
							var err error
							if r.Method == http.MethodPost {
								val, err = strconv.Atoi(r.FormValue("age"))
							} else {
								val, err = strconv.Atoi(r.URL.Query().Get("age"))
							}
							if err != nil {
								w.WriteHeader(http.StatusBadRequest)
								w.Write([]byte("{\"error\": \"age must be int\"}"))
								return
							}
							v.Age = val
						

						
							
							
						

						
						

						
							
								if v.Age < 0 {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"age must be >= 0\"}"))
									return
								}
							
						

						
							
								if v.Age > 128 {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"age must be <= 128\"}"))
									return
								}
							
						
						
					
				
			
			ctx := r.Context()
			res, err := this.Create(ctx, v)
			if err != nil {
				if apiErr, ok := err.(ApiError); ok {
					w.WriteHeader(apiErr.HTTPStatus)
					w.Write([]byte("{\"error\": \"" + apiErr.Error() + "\"}"))
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("{\"error\": \"" + err.Error() + "\"}"))
				}
				return
			}
			resMap := make(map[string]interface{})
			resMap["error"] = interface{}("")
			resMap["response"] = interface{}(res)
			jsonStr, err := json.Marshal(resMap)
			w.Write(jsonStr)	
		}
	

	
	func (this *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		
			if r.URL.Path == "/user/create" {
				this.handlerCreate(w, r)
				return
			}
		
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown method\"}"))
	}

	
		func (this *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
			
				if r.Method != "POST" {
					w.WriteHeader(http.StatusNotAcceptable)
					w.Write([]byte("{\"error\": \"bad method\"}"))
					return
				}
			
	
			if r.Method == http.MethodPost {
				r.ParseForm()
			}
			
			
				if r.Header.Get("X-Auth") != "100500" {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte("{\"error\": \"unauthorized\"}"))
					return
				}
			
			
				
				
			
				
				
					v := OtherCreateParams{}
					
						
							if r.Method == http.MethodPost {
								v.Username = r.FormValue("username")
							} else {
								v.Username = r.URL.Query().Get("username")
							}
						

						
							
								if v.Username == "" {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"username must me not empty\"}"))
									return
								}
							
							
						

						
						

						
							
								if len(v.Username) < 3 {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"username len must be >= 3\"}"))
									return
								}
							
						

						
						
					
						
							if r.Method == http.MethodPost {
								v.Name = r.FormValue("account_name")
							} else {
								v.Name = r.URL.Query().Get("account_name")
							}
						

						
							
							
						

						
						

						

						
						
					
						
							if r.Method == http.MethodPost {
								v.Class = r.FormValue("class")
							} else {
								v.Class = r.URL.Query().Get("class")
							}
						

						
							
							
								if v.Class == "" {
									v.Class = "warrior"
								}
							
						

						
						
							inEnum := false
							var errArr []string
							
							
							
							 	errArr = append(errArr, "warrior")
								if v.Class == "warrior" {
									inEnum = true
								}
							
							 	errArr = append(errArr, "sorcerer")
								if v.Class == "sorcerer" {
									inEnum = true
								}
							
							 	errArr = append(errArr, "rouge")
								if v.Class == "rouge" {
									inEnum = true
								}
								
							if !inEnum {
								w.WriteHeader(http.StatusBadRequest)
								w.Write([]byte("{\"error\": \"class must be one of ["+strings.Join(errArr, ", ")+"]\"}"))
								return
							}
						

						

						
						
					
						
							var val int
							var err error
							if r.Method == http.MethodPost {
								val, err = strconv.Atoi(r.FormValue("level"))
							} else {
								val, err = strconv.Atoi(r.URL.Query().Get("level"))
							}
							if err != nil {
								w.WriteHeader(http.StatusBadRequest)
								w.Write([]byte("{\"error\": \"level must be int\"}"))
								return
							}
							v.Level = val
						

						
							
							
						

						
						

						
							
								if v.Level < 1 {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"level must be >= 1\"}"))
									return
								}
							
						

						
							
								if v.Level > 50 {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"level must be <= 50\"}"))
									return
								}
							
						
						
					
				
			
			ctx := r.Context()
			res, err := this.Create(ctx, v)
			if err != nil {
				if apiErr, ok := err.(ApiError); ok {
					w.WriteHeader(apiErr.HTTPStatus)
					w.Write([]byte("{\"error\": \"" + apiErr.Error() + "\"}"))
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("{\"error\": \"" + err.Error() + "\"}"))
				}
				return
			}
			resMap := make(map[string]interface{})
			resMap["error"] = interface{}("")
			resMap["response"] = interface{}(res)
			jsonStr, err := json.Marshal(resMap)
			w.Write(jsonStr)	
		}
	
