package process_test

import "github.com/stretchr/testify/assert"

func (s *ProcessTestSuite) TestWorkingDirectory() {
	workdir, err := s.p.WorkingDirectory()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "/", workdir)

	err = s.p.ChangeDirectory("/.////../../a/b/../b/a/")
	assert.Nil(s.T(), err)

	workdir, err = s.p.WorkingDirectory()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "/a/b/a", workdir)
}
