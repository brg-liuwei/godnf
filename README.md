install:
    go get github.com/brg-liuwei/godnf

example:

    package main
    
    import (
    	"fmt"
    	dnf "godnf"
    )
    
    type attr struct {
    	h int
    	w int
    }
    
    func (this attr) ToString() string {
    	return fmt.Sprintf("{height: %d, width: %d}", this.h, this.w)
    }
    
    func (this attr) ToMap() map[string]interface{} {
    	return map[string]interface{}{
    		"height": this.h,
    		"width":  this.w,
    	}
    }
    
    func main() {
    	dnf.SetDebug(true)
    	dnf.SetHandler(dnf.NewHandler())
    	h := dnf.GetHandler()
    	var err error
    	err = h.AddDoc("ad0", "0", "(region in {ShangHai, Beijing} and age not in {3, 4})", attr{300, 250})
    	if err != nil {
    		panic(err)
    	}
    	err = h.AddDoc("ad1", "1", "(region in {ShenZhen, ShangHai}) or (age not in {4, 6})", attr{1080, 418})
    	if err != nil {
    		panic(err)
    	}
    	err = h.AddDoc("ad2", "2", "(region in {ShangHai, NanJing} and age not in {3, 5, 6})", attr{1080, 418})
    	if err != nil {
    		panic(err)
    	}
    	err = h.AddDoc("ad3", "3", "(region in {ChengDu, Beijing, WuHan}) or (age not in {4, 3})", attr{300, 250})
    	if err != nil {
    		panic(err)
    	}
    	err = h.AddDoc("ad4", "4", "(age not in {3, 4})", attr{300, 250})
    	if err != nil {
    		panic(err)
    	}
    
    	conds := []dnf.Cond{
    		{"region", "NanJing"},
    		{"age", "5"},
    	}

    	var docs []int
    	docs, err = h.Search(conds)
    	if err != nil {
    		panic(err)
    	}
    	fmt.Println("docs:", docs)
    }
