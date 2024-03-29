// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package hosercmd

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd(in *jlexer.Lexer, out *Start) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.Id = string(in.String())
		case "exe":
			out.ExeFile = string(in.String())
		case "argv":
			if in.IsNull() {
				in.Skip()
				out.Argv = nil
			} else {
				in.Delim('[')
				if out.Argv == nil {
					if !in.IsDelim(']') {
						out.Argv = make([]string, 0, 4)
					} else {
						out.Argv = []string{}
					}
				} else {
					out.Argv = (out.Argv)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Argv = append(out.Argv, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "ports":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				out.Ports = make(map[string]Port)
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v2 Port
					easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd1(in, &v2)
					(out.Ports)[key] = v2
					in.WantComma()
				}
				in.Delim('}')
			}
		default:
			in.AddError(&jlexer.LexerError{
				Offset: in.GetPos(),
				Reason: "unknown field",
				Data:   key,
			})
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd(out *jwriter.Writer, in Start) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.String(string(in.Id))
	}
	{
		const prefix string = ",\"exe\":"
		out.RawString(prefix)
		out.String(string(in.ExeFile))
	}
	{
		const prefix string = ",\"argv\":"
		out.RawString(prefix)
		if in.Argv == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v3, v4 := range in.Argv {
				if v3 > 0 {
					out.RawByte(',')
				}
				out.String(string(v4))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"ports\":"
		out.RawString(prefix)
		if in.Ports == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v5First := true
			for v5Name, v5Value := range in.Ports {
				if v5First {
					v5First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v5Name))
				out.RawByte(':')
				easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd1(out, v5Value)
			}
			out.RawByte('}')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Start) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Start) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Start) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Start) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd(l, v)
}
func easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd1(in *jlexer.Lexer, out *Port) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "dir":
			out.Dir = Dir(in.String())
		default:
			in.AddError(&jlexer.LexerError{
				Offset: in.GetPos(),
				Reason: "unknown field",
				Data:   key,
			})
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd1(out *jwriter.Writer, in Port) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"dir\":"
		out.RawString(prefix[1:])
		out.String(string(in.Dir))
	}
	out.RawByte('}')
}
func easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd2(in *jlexer.Lexer, out *Set) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.Id = string(in.String())
		case "read":
			out.Read = string(in.String())
		case "write":
			out.Write = string(in.String())
		case "text":
			out.Text = string(in.String())
		default:
			in.AddError(&jlexer.LexerError{
				Offset: in.GetPos(),
				Reason: "unknown field",
				Data:   key,
			})
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd2(out *jwriter.Writer, in Set) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.String(string(in.Id))
	}
	{
		const prefix string = ",\"read\":"
		out.RawString(prefix)
		out.String(string(in.Read))
	}
	{
		const prefix string = ",\"write\":"
		out.RawString(prefix)
		out.String(string(in.Write))
	}
	{
		const prefix string = ",\"text\":"
		out.RawString(prefix)
		out.String(string(in.Text))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Set) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Set) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Set) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Set) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd2(l, v)
}
func easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd3(in *jlexer.Lexer, out *Pipeline) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.Id = string(in.String())
		default:
			in.AddError(&jlexer.LexerError{
				Offset: in.GetPos(),
				Reason: "unknown field",
				Data:   key,
			})
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd3(out *jwriter.Writer, in Pipeline) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.String(string(in.Id))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Pipeline) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Pipeline) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Pipeline) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Pipeline) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd3(l, v)
}
func easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd4(in *jlexer.Lexer, out *Pipe) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "src":
			out.Src = string(in.String())
		case "dst":
			out.Dst = string(in.String())
		default:
			in.AddError(&jlexer.LexerError{
				Offset: in.GetPos(),
				Reason: "unknown field",
				Data:   key,
			})
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd4(out *jwriter.Writer, in Pipe) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"src\":"
		out.RawString(prefix[1:])
		out.String(string(in.Src))
	}
	{
		const prefix string = ",\"dst\":"
		out.RawString(prefix)
		out.String(string(in.Dst))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Pipe) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd4(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Pipe) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd4(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Pipe) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd4(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Pipe) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd4(l, v)
}
func easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd5(in *jlexer.Lexer, out *Exit) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "when":
			out.When = string(in.String())
		default:
			in.AddError(&jlexer.LexerError{
				Offset: in.GetPos(),
				Reason: "unknown field",
				Data:   key,
			})
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd5(out *jwriter.Writer, in Exit) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"when\":"
		out.RawString(prefix[1:])
		out.String(string(in.When))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Exit) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd5(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Exit) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF64fc67eEncodeGithubComHoserIoHoserRuntimeHosercmd5(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Exit) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd5(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Exit) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF64fc67eDecodeGithubComHoserIoHoserRuntimeHosercmd5(l, v)
}
