package pkg

import (
	"bufio"
	"crypto"
	"database/sql/driver"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"
)

type interfaceDelegate struct {
	vm       *VM
	receiver reflect.Value
	fun      *FuncDecl
}

// ============ io package (15 interfaces) ============

// io.Reader
func (d interfaceDelegate) Read(p []byte) (int, error) {
	d.vm.pushNewFrame(d.fun)
	d.vm.pushOperand(reflect.ValueOf(p))
	d.vm.pushOperand(d.receiver)
	call := CallExpr{args: []Expr{noExpr{}}}
	call.handleFuncDecl(d.vm, d.fun)
	return 0, io.EOF
}

// io.Writer
func (d interfaceDelegate) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// io.Closer
func (d interfaceDelegate) Close() error {
	return nil
}

// io.Seeker
func (d interfaceDelegate) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

// io.StringWriter
func (d interfaceDelegate) WriteString(s string) (n int, err error) {
	return len(s), nil
}

// io.ReaderAt
func (d interfaceDelegate) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, io.EOF
}

// io.WriterAt
func (d interfaceDelegate) WriteAt(p []byte, off int64) (n int, err error) {
	return len(p), nil
}

// io.ByteReader
func (d interfaceDelegate) ReadByte() (byte, error) {
	return 0, io.EOF
}

// io.ByteWriter
func (d interfaceDelegate) WriteByte(c byte) error {
	return nil
}

// io.RuneReader
func (d interfaceDelegate) ReadRune() (rune, int, error) {
	return 0, 0, io.EOF
}

// io.ByteScanner
func (d interfaceDelegate) UnreadByte() error {
	return nil
}

// io.RuneScanner
func (d interfaceDelegate) UnreadRune() error {
	return nil
}

// ============ fmt package (4 interfaces) ============

// fmt.Stringer
func (d interfaceDelegate) String() string {
	return "interfaceDelegate"
}

// error
func (d interfaceDelegate) Error() string {
	return "interfaceDelegate error"
}

// fmt.GoStringer
func (d interfaceDelegate) GoString() string {
	return "interfaceDelegate{}"
}

// fmt.Formatter
func (d interfaceDelegate) Format(s fmt.State, verb rune) {
	fmt.Fprintf(s, "interfaceDelegate")
}

// fmt.Scanner
func (d interfaceDelegate) Scan(state fmt.ScanState, verb rune) error {
	return nil
}

// ============ encoding package (8 interfaces) ============

// encoding.TextMarshaler
func (d interfaceDelegate) MarshalText() ([]byte, error) {
	return []byte(""), nil
}

// encoding.TextUnmarshaler
func (d interfaceDelegate) UnmarshalText(text []byte) error {
	return nil
}

// encoding.BinaryMarshaler
func (d interfaceDelegate) MarshalBinary() ([]byte, error) {
	return []byte(""), nil
}

// encoding.BinaryUnmarshaler
func (d interfaceDelegate) UnmarshalBinary(data []byte) error {
	return nil
}

// encoding/json.Marshaler
func (d interfaceDelegate) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

// encoding/json.Unmarshaler
func (d interfaceDelegate) UnmarshalJSON([]byte) error {
	return nil
}

// encoding/xml.Marshaler
func (d interfaceDelegate) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return nil
}

// encoding/xml.Unmarshaler
func (d interfaceDelegate) UnmarshalXML(e *xml.Decoder, start xml.StartElement) error {
	return nil
}

// ============ net package (8 interfaces) ============

// net.Addr
func (d interfaceDelegate) Network() string {
	return "unknown"
}

// net.Conn
func (d interfaceDelegate) LocalAddr() net.Addr {
	return d
}

// net.Conn (continued)
func (d interfaceDelegate) RemoteAddr() net.Addr {
	return d
}

// net.Conn (continued)
func (d interfaceDelegate) SetDeadline(t time.Time) error {
	return nil
}

// net.Conn (continued)
func (d interfaceDelegate) SetReadDeadline(t time.Time) error {
	return nil
}

// net.Conn (continued)
func (d interfaceDelegate) SetWriteDeadline(t time.Time) error {
	return nil
}

// net.Listener
func (d interfaceDelegate) Accept() (net.Conn, error) {
	return nil, io.EOF
}

// net.PacketConn
func (d interfaceDelegate) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return 0, nil, io.EOF
}

// net.PacketConn (continued)
func (d interfaceDelegate) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return len(p), nil
}

// ============ http package (12+ interfaces) ============

// net/http.Handler
func (d interfaceDelegate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

// net/http.ResponseWriter
func (d interfaceDelegate) Header() http.Header {
	return http.Header{}
}

// net/http.ResponseWriter (continued)
func (d interfaceDelegate) WriteHeader(statusCode int) {
}

// net/http.Flusher
func (d interfaceDelegate) Flush() {
}

// net/http.Hijacker
func (d interfaceDelegate) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, io.EOF
}

// net/http.RoundTripper
func (d interfaceDelegate) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, io.EOF
}

// net/http.CookieJar
func (d interfaceDelegate) SetCookies(u *url.URL, cookies []*http.Cookie) {
}

// net/http.CookieJar (continued)
func (d interfaceDelegate) Cookies(u *url.URL) []*http.Cookie {
	return nil
}

// net/http.Pusher
func (d interfaceDelegate) Push(target string, opts *http.PushOptions) error {
	return nil
}

// net/http.FileSystem
func (d interfaceDelegate) OpenFile(name string) (http.File, error) {
	return nil, io.EOF
}

// net/http.File
func (d interfaceDelegate) Readdir(count int) ([]os.FileInfo, error) {
	return nil, io.EOF
}

// ============ os package (3 interfaces) ============

// os.FileInfo
func (d interfaceDelegate) Name() string {
	return ""
}

// os.FileInfo (continued)
func (d interfaceDelegate) Size() int64 {
	return 0
}

// os.FileInfo (continued)
func (d interfaceDelegate) Mode() os.FileMode {
	return 0
}

// os.FileInfo (continued)
func (d interfaceDelegate) ModTime() time.Time {
	return time.Time{}
}

// os.FileInfo (continued)
func (d interfaceDelegate) IsDir() bool {
	return false
}

// os.DirEntry
func (d interfaceDelegate) Type() os.FileMode {
	return 0
}

// os.DirEntry (continued)
func (d interfaceDelegate) Info() (os.FileInfo, error) {
	return nil, io.EOF
}

// ============ bufio package (2 interfaces) ============

// bufio.Reader
func (d interfaceDelegate) Buffered() int {
	return 0
}

// ============ crypto package (6 interfaces) ============

// crypto.Hash
func (d interfaceDelegate) HashFunc() crypto.Hash {
	return 0
}

// crypto.Signer
func (d interfaceDelegate) Public() crypto.PublicKey {
	return nil
}

// crypto.Signer (continued)
func (d interfaceDelegate) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	return nil, io.EOF
}

// hash.Hash
func (d interfaceDelegate) Sum(b []byte) []byte {
	return nil
}

// hash.Hash (continued)
func (d interfaceDelegate) Reset() {
}

// hash.Hash32
func (d interfaceDelegate) Sum32() uint32 {
	return 0
}

// hash.Hash64
func (d interfaceDelegate) Sum64() uint64 {
	return 0
}

// ============ image package (5+ interfaces) ============

// image.Image
func (d interfaceDelegate) ColorModel() color.Model {
	return nil
}

// image.Image (continued)
func (d interfaceDelegate) Bounds() image.Rectangle {
	return image.Rectangle{}
}

// image.Image (continued)
func (d interfaceDelegate) At(x, y int) color.Color {
	return nil
}

// image.PalettedImage
func (d interfaceDelegate) ColorIndexAt(x, y int) uint8 {
	return 0
}

// image/draw.Image
func (d interfaceDelegate) Set(x, y int, c color.Color) {
}

// color.Color
func (d interfaceDelegate) RGBA() (r, g, b, a uint32) {
	return 0, 0, 0, 0
}

// color.Model
func (d interfaceDelegate) Convert(c color.Color) color.Color {
	return nil
}

// ============ sort package (2 interfaces) ============

// sort.Interface
func (d interfaceDelegate) Len() int {
	return 0
}

// sort.Interface (continued)
func (d interfaceDelegate) Less(i, j int) bool {
	return false
}

// sort.Interface (continued)
func (d interfaceDelegate) Swap(i, j int) {
}

// ============ sync package (1 interface) ============

// sync.Locker
func (d interfaceDelegate) Lock() {
}

// sync.Locker (continued)
func (d interfaceDelegate) Unlock() {
}

// ============ database/sql/driver package (6+ interfaces) ============

// database/sql/driver.Driver
func (d interfaceDelegate) OpenDB(name string) (driver.Conn, error) {
	return nil, io.EOF
}

// database/sql/driver.Conn
func (d interfaceDelegate) Prepare(query string) (driver.Stmt, error) {
	return nil, io.EOF
}

// database/sql/driver.Conn (continued)
func (d interfaceDelegate) Begin() (driver.Tx, error) {
	return nil, io.EOF
}

// database/sql/driver.Stmt
func (d interfaceDelegate) NumInput() int {
	return 0
}

// database/sql/driver.Stmt (continued)
func (d interfaceDelegate) ExecQuery(args []driver.Value) (driver.Result, error) {
	return nil, io.EOF
}

// database/sql/driver.Stmt (continued)
func (d interfaceDelegate) QueryStmt(args []driver.Value) (driver.Rows, error) {
	return nil, io.EOF
}

// database/sql/driver.Tx
func (d interfaceDelegate) Commit() error {
	return nil
}

// database/sql/driver.Tx (continued)
func (d interfaceDelegate) Rollback() error {
	return nil
}

// database/sql/driver.Rows
func (d interfaceDelegate) Columns() []string {
	return nil
}

// database/sql/driver.Rows (continued)
func (d interfaceDelegate) NextRow(dest []driver.Value) error {
	return io.EOF
}

// database/sql/driver.Result
func (d interfaceDelegate) LastInsertId() (int64, error) {
	return 0, io.EOF
}

// database/sql/driver.Result (continued)
func (d interfaceDelegate) RowsAffected() (int64, error) {
	return 0, io.EOF
}

// ============ Additional common interfaces ============

// bufio.Scanner (no methods, it's just a type)

// regexp.Regexp (no interfaces)

// text/scanner.Scanner (no interfaces)
