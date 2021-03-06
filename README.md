# gobatis

[![Appveyor Build status](https://ci.appveyor.com/api/projects/status/oou404q28phtxhwm?svg=true)](https://ci.appveyor.com/project/xfali/gobatis)

## 介绍

gobatis是一个golang的ORM框架，类似Java的Mybatis。支持直接执行sql语句以及动态sql。

建议配合[gobatis-cmd](https://github.com/xfali/gobatis-cmd)自动代码、sql生成工具使用。

支持的动态sql标签：

 标签 | 说明
:---: | :---
if | 动态 SQL 通常要做的事情是根据条件包含 where 子句的一部分。
where| where 元素只会在至少有一个子元素的条件返回 SQL 子句的情况下才去插入“WHERE”子句。而且，若语句的开头为“AND”或“OR”，where 元素也会将它们去除。 
set | set 元素会动态前置 SET 关键字，同时也会删掉无关的逗号。
include | 使用sql标签定义的语句替换。
choose<br>when<br>otherwise | 有时我们不想应用到所有的条件语句，而只想从中择其一项。针对这种情况，gobatis 提供了 choose 元素，它有点像switch 语句。
foreach | foreach 允许指定一个集合，声明可以在元素体内使用的集合项（item）和索引（index）变量。

## 待完成项

* 继续完善动态sql支持（trim）
* ~~性能优化：增加动态sql缓存~~
(已经实现，但测试发现性能提升很小，目前该功能被关闭)

## 使用


### 1、配置数据库，获得SessionManager

```
func InitDB() *gobatis.SessionManager {
    fac := factory.DefaultFactory{
        Host:     "localhost",
        Port:     3306,
        DBName:   "test",
        Username: "root",
        Password: "123",
        Charset:  "utf8",

        MaxConn:     1000,
        MaxIdleConn: 500,

        Log: logging.DefaultLogf,
    }
    fac.Init()
    return gobatis.NewSessionManager(&fac)
}
```

### 2、定义Model

使用tag（"xfield"）定义struct，tag指定数据库表中的column name。

```
type TestTable struct {
    //指定table name
    TestTable gobatis.ModelName "test_table"
    //指定表字段id
    Id        int64             `xfield:"id"`
    //指定表字段username
    Username  string            `xfield:"username"`
    //指定表字段password
    Password  string            `xfield:"password"`
}
```

### ~~3、注册Model~~

作用是提高执行速度，已变为非必要步骤，现在gobatis会自动缓存。
```
func init() {
    var model TestTable
    gobatis.RegisterModel(&model)
}
```

### 4、调用

```
func Run() {
    //初始化db并获得Session Manager
    mgr := InitDB()
    
    //获得session
    session := mgr.NewSession()
    
    ret := TestTable{}
    
    //使用session查询
    session.Select("select * from test_table where id = ${0}").Param(100).Result(&ret)
    
    fmt.printf("%v\n", ret)
}
```

### 5、说明

1. ${}表示直接替换，#{}防止sql注入
2. 与Mybatis类似，语句中${0}、${1}、${2}...${n} 对应的是Param方法中对应的不定参数，最终替换和调用底层Driver
3. Param方法接受简单类型的不定参数（string、int、time、float等）、struct、map，底层自动解析获得参数，用法为：

```
param := TestTable{Username:"test_user"}
ret := TestTable{}
session.Select("select * from test_table where username = #{TestTable.username}").Param(param).Result(&ret)
```

4. Param解析的参数规则（请务必按此规则对应SQL语句的占位参数）：
* 简单类型
  
  对应sql参数中的#{0}、#{1}...
  
* map类型

  对应sql参数中的#{key1}、#{key2}...
  
* struct类型
  
  对应sql参数中的#{StructName.Field1}、#{StructName.Field2}...
  

### 6、事务

使用
```
    mgr.NewSession().Tx(func(session *gobatis.Session) error {
        ret := 0
        session.Insert("insert_id").Param(testV).Result(&ret)
        
        t.Logf("ret %d\n", ret)
        
        session.Select("select_id").Param().Result(&testList)
        
        for _, v := range  testList {
            t.Logf("data: %v", v)
        }
        //commit
        return nil
    })
```
1. 当参数的func返回nil，则提交
2. 当参数的func返回非nil的错误，则回滚
3. 当参数的func内抛出panic，则回滚

### 7、xml

gobatis支持xml的sql解析及动态sql

1. 注册xml

```
gobatis.RegisterMapperData([]byte(main_xml))
```

或
    
```
gobatis.RegisterMapperFile(filePath)
```

2. xml示例

```
<mapper namespace="test_package.TestTable">
    <sql id="columns_id">`id`,`username`,`password`,`update_time`</sql>

    <select id="selectTestTable">
        SELECT <include refid="columns_id"> </include> FROM `TEST_TABLE`
        <where>
            <if test="{TestTable.id} != nil and {TestTable.id} != 0">AND `id` = #{TestTable.id} </if>
            <if test="{TestTable.username} != nil">AND `username` = #{TestTable.username} </if>
            <if test="{TestTable.password} != nil">AND `password` = #{TestTable.password} </if>
            <if test="{TestTable.update_time} != nil">AND `update_time` = #{TestTable.update_time} </if>
        </where>
    </select>

    <select id="selectTestTableCount">
        SELECT COUNT(*) FROM `TEST_TABLE`
        <where>
            <if test="{TestTable.id} != nil and {TestTable.id} != 0">AND `id` = #{TestTable.id} </if>
            <if test="{TestTable.username} != nil">AND `username` = #{TestTable.username} </if>
            <if test="{TestTable.password} != nil">AND `password` = #{TestTable.password} </if>
            <if test="{TestTable.update_time} != nil">AND `update_time` = #{TestTable.update_time} </if>
        </where>
    </select>

    <insert id="insertTestTable">
        INSERT INTO `TEST_TABLE` (`id`,`username`,`password`,`update_time`)
        VALUES(
        #{TestTable.id},
        #{TestTable.username},
        #{TestTable.password},
        #{TestTable.update_time}
        )
    </insert>

    <update id="updateTestTable">
        UPDATE `TEST_TABLE`
        <set>
            <if test="{TestTable.username} != nil"> `username` = #{TestTable.username} </if>
            <if test="{TestTable.password} != nil"> `password` = #{TestTable.password} </if>
            <if test="{TestTable.update_time} != nil"> `update_time` = #{TestTable.update_time} </if>
        </set>
        WHERE `id` = #{TestTable.id}
    </update>

    <delete id="deleteTestTable">
        DELETE FROM `TEST_TABLE`
        <where>
            <if test="{TestTable.id} != nil and {TestTable.id} != 0">AND `id` = #{TestTable.id} </if>
            <if test="{TestTable.username} != nil">AND `username` = #{TestTable.username} </if>
            <if test="{TestTable.password} != nil">AND `password` = #{TestTable.password} </if>
            <if test="{TestTable.update_time} != nil">AND `update_time` = #{TestTable.update_time} </if>
        </where>
    </delete>
</mapper>
```

### 8、gobatis-cmd生成文件使用示例

参考[cmd_test](https://github.com/xfali/gobatis/tree/master/test/cmd)

### 9、 SQL语句构建器

gobatis xml特性有非常强大的动态SQL生成方案，当需要在代码中嵌入SQL语句时，也可使用SQL语句构建器：
```
import "github.com/xfali/gobatis/builder"
```
```
    str := builder.Select("A.test1", "B.test2").
            Select("B.test3").
            From("test_a AS A").
            From("test_b AS B").
            Where("id = 1").
            And().
            Where("name=2").
            GroupBy("name").
            OrderBy("name").
            Desc().
            Offset(5).
            Limit(10).
            String()
    t.Log(str)
```
